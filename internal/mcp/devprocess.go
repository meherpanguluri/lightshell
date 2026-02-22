package mcp

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// MCPCommand is the JSON command sent from the MCP server to the dev process
// over the Unix socket.
type MCPCommand struct {
	ID       int    `json:"id"`
	Cmd      string `json:"cmd"`
	Delay    int    `json:"delay,omitempty"`    // for screenshot (ms to wait before capture)
	Lines    int    `json:"lines,omitempty"`    // for console (number of entries)
	Level    string `json:"level,omitempty"`    // for console (filter level)
	Clear    bool   `json:"clear,omitempty"`    // for console (clear after read)
	Selector string `json:"selector,omitempty"` // for dom (CSS selector)
	Depth    int    `json:"depth,omitempty"`    // for dom (traversal depth)
	Code     string `json:"code,omitempty"`     // for eval (JS code)
}

// MCPResponse is the JSON response from the dev process back to the MCP server.
type MCPResponse struct {
	ID      int              `json:"id"`
	Error   string           `json:"error,omitempty"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Image   string           `json:"image,omitempty"`
	Width   int              `json:"width,omitempty"`
	Height  int              `json:"height,omitempty"`
	Status  string           `json:"status,omitempty"`
	HTML    string           `json:"html,omitempty"`
	Entries []ConsoleEntry   `json:"entries,omitempty"`
}

// DevProcessManager manages the lightshell dev child process and communicates
// with it over a Unix domain socket.
type DevProcessManager struct {
	cmd        *exec.Cmd
	socketPath string
	conn       net.Conn
	reader     *bufio.Reader
	mu         sync.Mutex
	running    bool
	projectDir string
	nextID     atomic.Int64
	stderr     *limitedBuffer
	exitCh     chan error // signals when the child process exits
}

// limitedBuffer captures a limited amount of stderr output for error reporting.
type limitedBuffer struct {
	mu   sync.Mutex
	data []byte
	max  int
}

func newLimitedBuffer(max int) *limitedBuffer {
	return &limitedBuffer{data: make([]byte, 0, max), max: max}
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	remaining := b.max - len(b.data)
	if remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
		}
		b.data = append(b.data, p...)
	}
	return len(p), nil
}

func (b *limitedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.data)
}

// NewDevProcessManager creates a new dev process manager for the given project directory.
func NewDevProcessManager(projectDir string) *DevProcessManager {
	return &DevProcessManager{
		projectDir: projectDir,
		stderr:     newLimitedBuffer(4096),
	}
}

// Start launches the lightshell dev process with MCP socket support.
// It waits for the socket to become available and connects to it.
func (d *DevProcessManager) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// If already running, stop first
	if d.running {
		d.stopLocked()
	}

	// Generate socket path with random token to prevent prediction
	var token [8]byte
	if _, err := rand.Read(token[:]); err != nil {
		return fmt.Errorf("failed to generate socket token: %w", err)
	}
	d.socketPath = fmt.Sprintf("/tmp/lightshell-mcp-%d-%s.sock", os.Getpid(), hex.EncodeToString(token[:]))

	// Clean up any stale socket file
	os.Remove(d.socketPath)

	// Find the lightshell binary (self)
	selfPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find lightshell binary: %w", err)
	}

	// Spawn child: lightshell dev --mcp-socket <path>
	d.cmd = exec.Command(selfPath, "dev", "--mcp-socket", d.socketPath)
	d.cmd.Dir = d.projectDir
	d.cmd.Stdout = os.Stderr // Forward child stdout to our stderr for debugging
	d.stderr = newLimitedBuffer(4096)
	d.cmd.Stderr = d.stderr

	if err := d.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start dev process: %w", err)
	}

	// Start a goroutine to wait for the process to exit. This allows us to
	// detect early exits, since cmd.ProcessState is only populated after Wait().
	d.exitCh = make(chan error, 1)
	go func() {
		d.exitCh <- d.cmd.Wait()
	}()

	// Wait for the socket file to appear (poll every 100ms, up to 5 seconds)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(d.socketPath); err == nil {
			break
		}
		// Check if process exited early
		select {
		case <-d.exitCh:
			return fmt.Errorf("dev process exited early: %s", d.stderr.String())
		default:
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Verify socket exists
	if _, err := os.Stat(d.socketPath); err != nil {
		// Process may have failed to start — kill and wait via exitCh
		d.cmd.Process.Kill()
		<-d.exitCh
		return fmt.Errorf("dev process did not create socket within 5s: %s", d.stderr.String())
	}

	// Connect to the Unix socket
	conn, err := net.DialTimeout("unix", d.socketPath, 2*time.Second)
	if err != nil {
		d.cmd.Process.Kill()
		<-d.exitCh
		os.Remove(d.socketPath)
		return fmt.Errorf("failed to connect to dev process socket: %w", err)
	}

	d.conn = conn
	d.reader = bufio.NewReader(conn)
	d.running = true

	return nil
}

// Stop gracefully stops the dev process.
func (d *DevProcessManager) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.stopLocked()
}

// stopLocked stops the dev process. Must be called with d.mu held.
func (d *DevProcessManager) stopLocked() error {
	if !d.running {
		return nil
	}

	// Close the socket connection
	if d.conn != nil {
		d.conn.Close()
		d.conn = nil
		d.reader = nil
	}

	if d.cmd != nil && d.cmd.Process != nil {
		// Send SIGTERM
		d.cmd.Process.Signal(syscall.SIGTERM)

		// Wait up to 3 seconds for exit using the existing exitCh
		// (cmd.Wait() is already running in a goroutine from Start)
		if d.exitCh != nil {
			select {
			case <-d.exitCh:
				// Process exited cleanly
			case <-time.After(3 * time.Second):
				// Force kill
				d.cmd.Process.Kill()
				<-d.exitCh
			}
		}
	}

	// Clean up socket file
	if d.socketPath != "" {
		os.Remove(d.socketPath)
	}

	d.running = false
	d.cmd = nil

	return nil
}

// IsRunning returns whether the dev process is currently running.
func (d *DevProcessManager) IsRunning() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.running
}

// SendCommand sends a command to the dev process over the Unix socket and
// returns the response. Commands are serialized (one at a time).
func (d *DevProcessManager) SendCommand(cmd MCPCommand) (*MCPResponse, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return nil, fmt.Errorf("dev process is not running")
	}

	// Assign a command ID
	cmd.ID = int(d.nextID.Add(1))

	// Marshal command to JSON + newline
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}
	data = append(data, '\n')

	// Set write deadline
	d.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	if _, err := d.conn.Write(data); err != nil {
		// Connection may be broken; mark as not running
		d.running = false
		return nil, fmt.Errorf("failed to write command: %w", err)
	}

	// Set read deadline (longer for screenshot/eval which may take time)
	timeout := 10 * time.Second
	if cmd.Cmd == "screenshot" {
		timeout = 15 * time.Second
	}
	d.conn.SetReadDeadline(time.Now().Add(timeout))

	// Read response line (newline-delimited)
	line, err := d.reader.ReadBytes('\n')
	if err != nil {
		d.running = false
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resp MCPResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Error != "" {
		return &resp, fmt.Errorf("%s", resp.Error)
	}

	return &resp, nil
}

// Cleanup is called on MCP server shutdown. It stops the dev process and
// cleans up all resources.
func (d *DevProcessManager) Cleanup() {
	d.Stop()
}

// SocketPath returns the path of the Unix domain socket used for communication.
func (d *DevProcessManager) SocketPath() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.socketPath
}
