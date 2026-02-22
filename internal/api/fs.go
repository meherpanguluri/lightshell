package api

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/lightshell-dev/lightshell/internal/ipc"
	"github.com/lightshell-dev/lightshell/internal/security"
)

// RegisterFS registers file system API handlers with security checks.
func RegisterFS(router *ipc.Router, policy *security.Policy) {
	router.Handle("fs.readFile", func(params json.RawMessage) (any, error) {
		if err := policy.Check(security.PermFS); err != nil {
			return nil, err
		}
		var p struct {
			Path     string `json:"path"`
			Encoding string `json:"encoding"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		if err := policy.CheckFSRead(p.Path); err != nil {
			return nil, err
		}
		data, err := os.ReadFile(p.Path)
		if err != nil {
			return nil, err
		}
		switch p.Encoding {
		case "base64", "binary":
			return base64.StdEncoding.EncodeToString(data), nil
		default:
			return string(data), nil
		}
	})

	router.Handle("fs.writeFile", func(params json.RawMessage) (any, error) {
		if err := policy.Check(security.PermFS); err != nil {
			return nil, err
		}
		var p struct {
			Path string `json:"path"`
			Data string `json:"data"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		if err := policy.CheckFSWrite(p.Path); err != nil {
			return nil, err
		}
		dir := filepath.Dir(p.Path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		return nil, os.WriteFile(p.Path, []byte(p.Data), 0o644)
	})

	router.Handle("fs.readDir", func(params json.RawMessage) (any, error) {
		if err := policy.Check(security.PermFS); err != nil {
			return nil, err
		}
		var p struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		if err := policy.CheckFSRead(p.Path); err != nil {
			return nil, err
		}
		entries, err := os.ReadDir(p.Path)
		if err != nil {
			return nil, err
		}
		result := make([]map[string]any, 0, len(entries))
		for _, e := range entries {
			info, _ := e.Info()
			size := int64(0)
			if info != nil {
				size = info.Size()
			}
			result = append(result, map[string]any{
				"name":  e.Name(),
				"isDir": e.IsDir(),
				"size":  size,
			})
		}
		return result, nil
	})

	router.Handle("fs.exists", func(params json.RawMessage) (any, error) {
		if err := policy.Check(security.PermFS); err != nil {
			return nil, err
		}
		var p struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		if err := policy.CheckFSRead(p.Path); err != nil {
			return nil, err
		}
		_, err := os.Stat(p.Path)
		return err == nil, nil
	})

	router.Handle("fs.stat", func(params json.RawMessage) (any, error) {
		if err := policy.Check(security.PermFS); err != nil {
			return nil, err
		}
		var p struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		if err := policy.CheckFSRead(p.Path); err != nil {
			return nil, err
		}
		info, err := os.Stat(p.Path)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"name":    info.Name(),
			"size":    info.Size(),
			"isDir":   info.IsDir(),
			"modTime": info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
			"mode":    info.Mode().String(),
		}, nil
	})

	router.Handle("fs.mkdir", func(params json.RawMessage) (any, error) {
		if err := policy.Check(security.PermFS); err != nil {
			return nil, err
		}
		var p struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		if err := policy.CheckFSWrite(p.Path); err != nil {
			return nil, err
		}
		return nil, os.MkdirAll(p.Path, 0o755)
	})

	router.Handle("fs.remove", func(params json.RawMessage) (any, error) {
		if err := policy.Check(security.PermFS); err != nil {
			return nil, err
		}
		var p struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		if err := policy.CheckFSWrite(p.Path); err != nil {
			return nil, err
		}
		return nil, os.Remove(p.Path)
	})
}
