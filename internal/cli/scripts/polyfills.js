(()=>{
const L=navigator.platform.includes('Linux')
document.documentElement.classList.add(L?'platform-linux':'platform-darwin')
if(typeof structuredClone==='undefined'){
window.structuredClone=(o,t)=>{
if(t?.transfer?.length)throw new DOMException('Transfer not supported','DataCloneError')
return JSON.parse(JSON.stringify(o))
}}
if(!Array.prototype.group){
Array.prototype.group=function(f){
return this.reduce((a,v,i)=>{(a[f(v,i,this)]??=[]).push(v);return a},{})
}}
if(!Promise.withResolvers){
Promise.withResolvers=function(){let r,j;const p=new Promise((a,b)=>{r=a;j=b});return{promise:p,resolve:r,reject:j}}
}
if(!Set.prototype.union){
const S=Set.prototype
S.union=function(o){const r=new Set(this);for(const v of o)r.add(v);return r}
S.intersection=function(o){const r=new Set;for(const v of this)if(o.has(v))r.add(v);return r}
S.difference=function(o){const r=new Set;for(const v of this)if(!o.has(v))r.add(v);return r}
S.symmetricDifference=function(o){const r=new Set(this);for(const v of o)r.has(v)?r.delete(v):r.add(v);return r}
S.isSubsetOf=function(o){for(const v of this)if(!o.has(v))return!1;return!0}
S.isSupersetOf=function(o){for(const v of o)if(!this.has(v))return!1;return!0}
S.isDisjointFrom=function(o){for(const v of this)if(o.has(v))return!1;return!0}
}
if(!Object.groupBy){
Object.groupBy=(t,f)=>{const r=Object.create(null);let i=0;for(const v of t){(r[f(v,i++)]??=[]).push(v)}return r}
}
if(typeof Intl!=='undefined'&&!Intl.Segmenter)console.warn('[LightShell] Intl.Segmenter unavailable')
if(L&&typeof CSS!=='undefined'&&CSS.supports&&!CSS.supports('backdrop-filter','blur(1px)')){
const s=document.createElement('style')
s.textContent='[style*="backdrop-filter"]{background-color:rgba(255,255,255,.9)!important}'
document.head.appendChild(s)
}
})()
