// Code generated by go-bindata. DO NOT EDIT.
// sources:
// 001_init.down.sql (20B)
// 001_init.up.sql (477B)
// 002_commands.down.sql (21B)
// 002_commands.up.sql (479B)
// 003_add_fields.down.sql (0)
// 003_add_fields.up.sql (352B)
// 004_add_timeout.down.sql (0)
// 004_add_timeout.up.sql (144B)

package library

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %w", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes  []byte
	info   os.FileInfo
	digest [sha256.Size]byte
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var __001_initDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\x28\x4e\x2e\xca\x2c\x28\x29\xb6\xe6\x02\x04\x00\x00\xff\xff\x24\x6d\x54\xc3\x14\x00\x00\x00")

func _001_initDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__001_initDownSql,
		"001_init.down.sql",
	)
}

func _001_initDownSql() (*asset, error) {
	bytes, err := _001_initDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "001_init.down.sql", size: 20, mode: os.FileMode(0644), modTime: time.Unix(1685339921, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xd4, 0x6f, 0x27, 0xe0, 0x17, 0x7c, 0x7b, 0x1d, 0xfe, 0x69, 0xdf, 0x43, 0x5e, 0x10, 0x9a, 0x2, 0x68, 0xfa, 0x83, 0xd7, 0x3d, 0x2c, 0x19, 0xa4, 0x78, 0x4b, 0x12, 0xff, 0x32, 0xf7, 0x16, 0xa5}}
	return a, nil
}

var __001_initUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x90\xcb\x6a\xc3\x30\x10\x45\xf7\xfa\x8a\x4b\x56\x31\xd4\xd0\xae\xb3\x32\x89\x28\xa6\xa9\xd2\xba\x32\x24\x2b\x47\xb1\x26\x60\x70\xe5\x54\x1a\x51\xfa\xf7\x25\x96\x29\x69\xe9\x4b\x3b\xcd\x9d\x73\xa4\x99\x3c\x47\xfe\xcb\x11\x79\x0e\x6d\x0e\x3d\x21\xb0\x8f\x2d\x47\x4f\x38\x0e\x1e\xa1\xf5\xdd\x89\x83\xf8\x0b\x6f\x3d\x19\x26\x70\x52\x4c\xd0\x5c\x00\x40\x67\xa1\xe5\x56\xe3\xa1\x2a\xef\x8b\x6a\x87\x3b\xb9\x83\xda\x68\xa8\x7a\xbd\xbe\x1a\x3b\x9c\x79\xa6\xd4\xe3\x06\x86\x8b\x7d\x9f\xea\x49\x6a\x1b\xc3\x58\x15\x5a\xfe\x90\x1e\xde\xbe\x63\x3b\xc7\xe4\x4f\x9e\x98\xfc\x18\x4f\xd5\xd0\x84\x68\x07\x94\x4a\xcb\x5b\x59\xcd\x6f\x32\x58\x3a\x9a\xd8\x33\xae\xbf\xea\x5f\xed\x05\x98\x46\xfa\xfc\x90\xc8\x16\x42\x2c\x2b\x79\xfe\x5a\xa9\x56\x72\x8b\xd9\x79\x92\xd9\x08\x6c\x14\xf6\xd3\x1e\xf6\x48\x8b\x48\x29\x8a\xa7\xe5\x78\xbd\xa0\x6b\x55\x3e\xd6\x1f\x92\xe8\xba\x97\x48\xcd\x7f\x5d\xd9\x42\xbc\x07\x00\x00\xff\xff\xe9\x96\x0a\x2f\xdd\x01\x00\x00")

func _001_initUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__001_initUpSql,
		"001_init.up.sql",
	)
}

func _001_initUpSql() (*asset, error) {
	bytes, err := _001_initUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "001_init.up.sql", size: 477, mode: os.FileMode(0644), modTime: time.Unix(1685339921, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x5a, 0x16, 0xa5, 0xfe, 0x93, 0x54, 0x6b, 0x8d, 0x6f, 0x14, 0xc, 0xfe, 0xfe, 0xed, 0xe9, 0xb4, 0x6, 0xe1, 0xdb, 0xb7, 0xc5, 0x1e, 0x98, 0xbb, 0x90, 0xef, 0x8d, 0x8e, 0xfb, 0xf0, 0x99, 0xdc}}
	return a, nil
}

var __002_commandsDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\x48\xce\xcf\xcd\x4d\xcc\x4b\x29\xb6\xe6\x02\x04\x00\x00\xff\xff\xb6\x29\x99\x09\x15\x00\x00\x00")

func _002_commandsDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__002_commandsDownSql,
		"002_commands.down.sql",
	)
}

func _002_commandsDownSql() (*asset, error) {
	bytes, err := _002_commandsDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "002_commands.down.sql", size: 21, mode: os.FileMode(0644), modTime: time.Unix(1685339921, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x91, 0x78, 0x3e, 0xea, 0xb, 0x4e, 0x3b, 0xc2, 0x71, 0x25, 0xf3, 0x3, 0x52, 0xdb, 0x71, 0x67, 0xf, 0x61, 0xde, 0xb, 0xa2, 0x2b, 0xdd, 0x78, 0x8c, 0xec, 0x6d, 0x35, 0x97, 0xbc, 0xf5, 0xf1}}
	return a, nil
}

var __002_commandsUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x8f\xc1\x4e\xc4\x20\x14\x45\xf7\x7c\xc5\x4d\x57\x4e\x22\x5f\x30\xab\x66\x86\x45\xe3\xc8\x68\xa5\xc9\xcc\xaa\x43\x01\x93\x26\x85\x2a\x85\x85\x7f\x6f\x0a\xa9\xd1\xa4\x56\xd9\x71\xef\x3b\x27\xef\x51\x0a\xba\xf1\x08\xa5\x10\xb2\x1b\x0c\xa6\xe0\xa3\x0a\xd1\x1b\xbc\x8e\x1e\x6a\xb4\x56\x3a\x3d\x91\xbf\x78\xe5\x8d\x0c\x06\x21\x39\xbe\xa8\x3b\x02\x00\xbd\x86\x60\x17\x81\xa7\xba\x7a\x2c\xeb\x2b\x1e\xd8\x15\xfc\x2c\xc0\x9b\xd3\xe9\x3e\x4d\x38\x69\x4d\x9e\x71\x63\x80\x8b\xc3\x90\xf3\x6c\xd5\xad\x0c\x38\x96\x82\xfd\xd2\x76\x1f\x6b\x6c\x7c\xd3\x1b\xec\xd2\xae\xb3\xca\xea\x9f\x31\xd9\xed\x09\x39\xd4\x6c\x16\x55\xfc\xc8\x2e\x28\x96\x23\xdb\x76\xde\xbe\x48\xdc\x99\xe3\xb6\xe4\x37\xe4\xeb\x8b\x54\xa3\x7c\x39\xa4\xef\x37\x51\xc3\xab\xe7\x66\xc5\x17\x5d\xff\x1e\xcd\xff\xb5\xbb\x3d\xf9\x0c\x00\x00\xff\xff\x78\x59\x4f\x68\xdf\x01\x00\x00")

func _002_commandsUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__002_commandsUpSql,
		"002_commands.up.sql",
	)
}

func _002_commandsUpSql() (*asset, error) {
	bytes, err := _002_commandsUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "002_commands.up.sql", size: 479, mode: os.FileMode(0644), modTime: time.Unix(1685339921, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xcf, 0x6d, 0x8b, 0x84, 0x21, 0xcb, 0xa5, 0x27, 0x36, 0x58, 0xe5, 0xe1, 0x34, 0x91, 0xc2, 0xeb, 0xc8, 0x34, 0x5c, 0x74, 0x87, 0x5c, 0xab, 0x57, 0xdc, 0xfd, 0x4, 0x59, 0xb4, 0xb2, 0x2b, 0x7d}}
	return a, nil
}

var __003_add_fieldsDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x01\x00\x00\xff\xff\x00\x00\x00\x00\x00\x00\x00\x00")

func _003_add_fieldsDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__003_add_fieldsDownSql,
		"003_add_fields.down.sql",
	)
}

func _003_add_fieldsDownSql() (*asset, error) {
	bytes, err := _003_add_fieldsDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "003_add_fields.down.sql", size: 0, mode: os.FileMode(0644), modTime: time.Unix(1685339921, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14, 0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c, 0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55}}
	return a, nil
}

var __003_add_fieldsUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\x50\x2a\x4e\x2e\xca\x2c\x28\x29\x56\x52\x70\x74\x71\x51\x70\xf6\xf7\x09\xf5\xf5\x53\x50\x2a\x2d\x48\x49\x2c\x49\x4d\x89\x4f\xaa\x54\x52\x08\x71\x8d\x08\x51\xf0\xf3\x0f\x51\xf0\x0b\xf5\xf1\x51\x70\x71\x75\x73\x0c\xf5\x09\x51\x50\x57\xb7\xe6\x22\xde\x9c\xc4\x12\x25\x05\x17\xc7\x10\x57\x4c\x73\xf2\xf2\xcb\xad\xb9\x42\x03\xc0\x92\x50\x23\x14\x82\x5d\x43\x14\x10\x3a\x15\x6c\x15\x92\x8b\x52\xa1\x1c\x1d\x05\x84\xd3\x90\x24\x92\x2a\xad\xb9\x50\x9d\x93\x9c\x9f\x9b\x9b\x98\x97\x82\xe6\x9e\x92\xc4\xf4\x62\x9c\x3e\x8a\x8e\x25\xca\x4f\x84\xcd\x00\x04\x00\x00\xff\xff\x1f\x29\xfb\x34\x60\x01\x00\x00")

func _003_add_fieldsUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__003_add_fieldsUpSql,
		"003_add_fields.up.sql",
	)
}

func _003_add_fieldsUpSql() (*asset, error) {
	bytes, err := _003_add_fieldsUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "003_add_fields.up.sql", size: 352, mode: os.FileMode(0644), modTime: time.Unix(1685339921, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x32, 0xd9, 0xf, 0x50, 0x98, 0x53, 0xaa, 0xeb, 0x26, 0xe7, 0x15, 0x5c, 0xe6, 0xa5, 0x6d, 0x9d, 0x31, 0xd1, 0xaf, 0xe4, 0x2d, 0x59, 0x85, 0x9, 0x1, 0xb7, 0x7b, 0xc6, 0xa5, 0x56, 0xfe, 0x27}}
	return a, nil
}

var __004_add_timeoutDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x01\x00\x00\xff\xff\x00\x00\x00\x00\x00\x00\x00\x00")

func _004_add_timeoutDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__004_add_timeoutDownSql,
		"004_add_timeout.down.sql",
	)
}

func _004_add_timeoutDownSql() (*asset, error) {
	bytes, err := _004_add_timeoutDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "004_add_timeout.down.sql", size: 0, mode: os.FileMode(0644), modTime: time.Unix(1685339921, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14, 0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c, 0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55}}
	return a, nil
}

var __004_add_timeoutUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\x50\x4a\xce\xcf\xcd\x4d\xcc\x4b\x29\x56\x52\x70\x74\x71\x51\x70\xf6\xf7\x09\xf5\xf5\x53\x50\x2a\xc9\xcc\x4d\xcd\x2f\x2d\x89\x2f\x4e\x4d\x56\x52\xf0\xf4\x0b\x51\xf0\xf3\x0f\x51\xf0\x0b\xf5\xf1\x51\x70\x71\x75\x73\x0c\xf5\x09\x51\x30\x33\xb0\xe6\x42\x31\xa8\x38\xb9\x28\xb3\xa0\x84\x1c\x73\x00\x01\x00\x00\xff\xff\xee\x35\xd6\x0b\x90\x00\x00\x00")

func _004_add_timeoutUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__004_add_timeoutUpSql,
		"004_add_timeout.up.sql",
	)
}

func _004_add_timeoutUpSql() (*asset, error) {
	bytes, err := _004_add_timeoutUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "004_add_timeout.up.sql", size: 144, mode: os.FileMode(0644), modTime: time.Unix(1685339921, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x13, 0x5e, 0x9, 0xc6, 0xa9, 0xf1, 0x72, 0xa4, 0x1f, 0x1f, 0x62, 0x50, 0xe1, 0x1b, 0xc6, 0x9d, 0xfc, 0x52, 0x5f, 0x49, 0x6c, 0x2f, 0x4e, 0x4, 0xe, 0x94, 0x8, 0xa6, 0x4e, 0xef, 0x39, 0x28}}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetString returns the asset contents as a string (instead of a []byte).
func AssetString(name string) (string, error) {
	data, err := Asset(name)
	return string(data), err
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// MustAssetString is like AssetString but panics when Asset would return an
// error. It simplifies safe initialization of global variables.
func MustAssetString(name string) string {
	return string(MustAsset(name))
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetDigest returns the digest of the file with the given name. It returns an
// error if the asset could not be found or the digest could not be loaded.
func AssetDigest(name string) ([sha256.Size]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return [sha256.Size]byte{}, fmt.Errorf("AssetDigest %s can't read by error: %v", name, err)
		}
		return a.digest, nil
	}
	return [sha256.Size]byte{}, fmt.Errorf("AssetDigest %s not found", name)
}

// Digests returns a map of all known files and their checksums.
func Digests() (map[string][sha256.Size]byte, error) {
	mp := make(map[string][sha256.Size]byte, len(_bindata))
	for name := range _bindata {
		a, err := _bindata[name]()
		if err != nil {
			return nil, err
		}
		mp[name] = a.digest
	}
	return mp, nil
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"001_init.down.sql":        _001_initDownSql,
	"001_init.up.sql":          _001_initUpSql,
	"002_commands.down.sql":    _002_commandsDownSql,
	"002_commands.up.sql":      _002_commandsUpSql,
	"003_add_fields.down.sql":  _003_add_fieldsDownSql,
	"003_add_fields.up.sql":    _003_add_fieldsUpSql,
	"004_add_timeout.down.sql": _004_add_timeoutDownSql,
	"004_add_timeout.up.sql":   _004_add_timeoutUpSql,
}

// AssetDebug is true if the assets were built with the debug flag enabled.
const AssetDebug = false

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//
//	data/
//	  foo.txt
//	  img/
//	    a.png
//	    b.png
//
// then AssetDir("data") would return []string{"foo.txt", "img"},
// AssetDir("data/img") would return []string{"a.png", "b.png"},
// AssetDir("foo.txt") and AssetDir("notexist") would return an error, and
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		canonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(canonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"001_init.down.sql":        {_001_initDownSql, map[string]*bintree{}},
	"001_init.up.sql":          {_001_initUpSql, map[string]*bintree{}},
	"002_commands.down.sql":    {_002_commandsDownSql, map[string]*bintree{}},
	"002_commands.up.sql":      {_002_commandsUpSql, map[string]*bintree{}},
	"003_add_fields.down.sql":  {_003_add_fieldsDownSql, map[string]*bintree{}},
	"003_add_fields.up.sql":    {_003_add_fieldsUpSql, map[string]*bintree{}},
	"004_add_timeout.down.sql": {_004_add_timeoutDownSql, map[string]*bintree{}},
	"004_add_timeout.up.sql":   {_004_add_timeoutUpSql, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory.
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = os.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	return os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
}

// RestoreAssets restores an asset under the given directory recursively.
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(canonicalName, "/")...)...)
}