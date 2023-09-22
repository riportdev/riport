// Code generated by go-bindata. DO NOT EDIT.
// sources:
// 001_init.down.sql (21B)
// 001_init.up.sql (301B)
// 002_plural_and_name.down.sql (120B)
// 002_plural_and_name.up.sql (169B)
// 003_init.down.sql (57B)
// 003_init.up.sql (513B)

package api_token

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

var __001_initDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\x48\x2c\xc8\x8c\x2f\xc9\xcf\x4e\xcd\xb3\x06\x04\x00\x00\xff\xff\x8d\x12\x02\x7d\x15\x00\x00\x00")

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

	info := bindataFileInfo{name: "001_init.down.sql", size: 21, mode: os.FileMode(0644), modTime: time.Unix(1685339920, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xe1, 0x69, 0x27, 0xb6, 0x18, 0xa6, 0xaf, 0x8f, 0x84, 0xe3, 0x84, 0x6b, 0xb5, 0x2a, 0x1, 0x4b, 0xae, 0x8d, 0x91, 0x61, 0x91, 0x6d, 0x10, 0x5, 0xb3, 0xc6, 0xf2, 0xd4, 0xc5, 0xbb, 0x48, 0x8}}
	return a, nil
}

var __001_initUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x8f\x41\x6a\xc3\x30\x10\x45\xf7\x39\xc5\xef\x2a\x31\xf8\x06\xa5\x0b\x55\x9e\x12\x11\xd9\x0e\xea\x88\x34\x2b\x23\xd2\x29\x98\xd2\x44\xc8\x2e\xe4\xf8\xa5\x91\xa9\x5d\xc8\x52\x3c\xbd\xe1\x3f\xed\x48\x31\x81\xd5\xb3\x25\x84\xd8\x77\xe3\xe5\x53\xce\xd8\xac\x00\xe0\x7b\x90\x74\x0e\x5f\x02\xa6\x37\x46\xd3\x32\x1a\x6f\x2d\xf4\x96\xf4\x0e\x9b\x3f\xfa\xf0\x84\xf5\xba\x28\x6f\x4a\x4c\xf2\xd1\x5f\xef\x0b\x13\x5b\x7e\x3f\x25\x09\xa3\xbc\x77\x61\x44\xa5\x98\xd8\xd4\x34\x6b\x15\xbd\x28\x6f\x19\xda\x3b\x47\x0d\x77\xbf\xf4\x95\x55\xbd\x2f\x71\x93\xe5\x1a\xfb\x24\xc3\x52\xce\x57\x87\xd3\x25\xe6\xd1\xf9\x9d\x9b\xfe\x6d\xca\x60\xef\x4c\xad\xdc\x11\x3b\x3a\xce\x3d\xe5\x14\x51\xac\x0a\x1c\x0c\x6f\x5b\xcf\x70\xed\xc1\x54\x8f\x3f\x01\x00\x00\xff\xff\x2c\xf2\x75\x27\x2d\x01\x00\x00")

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

	info := bindataFileInfo{name: "001_init.up.sql", size: 301, mode: os.FileMode(0644), modTime: time.Unix(1685339920, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xdc, 0x65, 0xa2, 0x1f, 0xba, 0xc2, 0xa5, 0x5e, 0x14, 0x1d, 0xf9, 0x67, 0xd7, 0xa9, 0x27, 0xe0, 0x83, 0xc7, 0x85, 0x99, 0xb2, 0x7c, 0x7b, 0x4d, 0x69, 0xcc, 0x32, 0xaa, 0x86, 0x42, 0x54, 0x2c}}
	return a, nil
}

var __002_plural_and_nameDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\xf0\xf4\x73\x71\x8d\x50\x48\x2c\xc8\x8c\x2f\xc9\xcf\x4e\xcd\x2b\x8e\x2f\xcd\xcb\x2c\x2c\x4d\x8d\xcf\x4b\xcc\x4d\xb5\xe6\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x45\x52\xa3\x90\x52\x94\x5f\xa0\xe0\xec\xef\x13\xea\xeb\xa7\x80\x57\x61\x90\xab\x9f\xa3\xaf\xab\x42\x88\x3f\x42\xd0\x9a\x0b\x10\x00\x00\xff\xff\x93\xa1\xea\x76\x78\x00\x00\x00")

func _002_plural_and_nameDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__002_plural_and_nameDownSql,
		"002_plural_and_name.down.sql",
	)
}

func _002_plural_and_nameDownSql() (*asset, error) {
	bytes, err := _002_plural_and_nameDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "002_plural_and_name.down.sql", size: 120, mode: os.FileMode(0644), modTime: time.Unix(1685339920, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x44, 0x95, 0x1c, 0xcb, 0xfc, 0xe8, 0x14, 0xa4, 0xd, 0x4, 0x2, 0x9f, 0x28, 0xba, 0x6e, 0xdd, 0xe5, 0x1c, 0xc5, 0x8c, 0x53, 0x30, 0xba, 0x33, 0x6f, 0x8e, 0xef, 0x82, 0x8, 0xd7, 0xe5, 0x65}}
	return a, nil
}

var __002_plural_and_nameUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x8c\xbb\x0a\x02\x31\x10\x45\xfb\x7c\xc5\x2d\x15\xfc\x83\x54\x51\xa7\x58\x88\x13\x0c\x13\xd8\x2e\x04\x76\x8a\x45\x8c\x8f\x98\xff\x97\xb5\x71\x8b\x6d\xef\x3d\xe7\x38\x2f\x14\x21\xee\xe8\x09\xe5\x39\xe7\xcf\xe3\xa6\x15\x91\xd8\x5d\x08\x12\xfe\x5b\xb3\x66\x93\x6d\x28\xd3\x84\x5a\xee\x0a\xa1\x51\xc0\x41\xc0\xc9\x7b\x6b\x4e\x91\x9c\x10\x12\x0f\xd7\x44\x18\xf8\x4c\xe3\x4a\xcb\xbd\xce\xaf\xae\x79\x31\x0d\x00\x04\x5e\x47\x77\xbd\xe9\x7b\xf9\x0e\xbf\xf6\xde\x7e\x03\x00\x00\xff\xff\xcc\x54\x2b\x61\xa9\x00\x00\x00")

func _002_plural_and_nameUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__002_plural_and_nameUpSql,
		"002_plural_and_name.up.sql",
	)
}

func _002_plural_and_nameUpSql() (*asset, error) {
	bytes, err := _002_plural_and_nameUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "002_plural_and_name.up.sql", size: 169, mode: os.FileMode(0644), modTime: time.Unix(1685339920, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xcc, 0x88, 0x2c, 0x5c, 0x18, 0x80, 0x1e, 0x9e, 0x81, 0x82, 0xb7, 0xb5, 0xf2, 0x9, 0x5, 0x5c, 0x76, 0xcb, 0xdb, 0x73, 0xc2, 0x95, 0xc1, 0xb7, 0xae, 0x6d, 0x4b, 0x1, 0x65, 0xf2, 0x3b, 0x57}}
	return a, nil
}

var __003_initDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\x48\x2c\xc8\x8c\x2f\xc9\xcf\x4e\xcd\x2b\xb6\xe6\x02\x0b\x7b\xfa\xb9\xb8\x46\x20\x09\xc7\x97\xe6\x65\x16\x96\xa6\xc6\xe7\x25\xe6\xa6\x5a\x03\x02\x00\x00\xff\xff\xc6\x50\x05\x19\x39\x00\x00\x00")

func _003_initDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__003_initDownSql,
		"003_init.down.sql",
	)
}

func _003_initDownSql() (*asset, error) {
	bytes, err := _003_initDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "003_init.down.sql", size: 57, mode: os.FileMode(0644), modTime: time.Unix(1685339920, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x97, 0x6b, 0x57, 0x42, 0xe3, 0xba, 0xee, 0x4b, 0x16, 0xcf, 0xc7, 0xff, 0xed, 0xbf, 0xaa, 0x66, 0xdf, 0x59, 0xa0, 0x3e, 0x14, 0x21, 0x74, 0xe8, 0x9f, 0x6c, 0x75, 0x22, 0x34, 0xa2, 0x51, 0x81}}
	return a, nil
}

var __003_initUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x91\xcf\x6e\x83\x30\x0c\xc6\xef\x79\x0a\xef\xd4\x22\xf5\x0d\xd0\x0e\x19\xb8\x6a\x54\x08\x2c\x38\x2a\x3d\x21\xd4\x65\x12\x9a\x46\x19\x7f\xa4\x3e\xfe\x44\xd2\x2d\x54\xe5\x82\x44\xbe\xef\x67\xfb\xb3\x63\x95\xe5\x40\xfc\x2d\x41\x10\x7b\xc0\x52\x14\x54\x40\xdd\x35\xd5\x78\xfd\x32\xed\x10\x32\x6b\x10\x32\xc6\x72\xd5\x50\x4d\x6d\xf3\x33\x99\xaa\xad\xbf\x4d\xc8\x58\xa4\x90\x13\xfa\x82\x32\xa3\x67\x06\xb6\x0c\x00\x60\x1a\x4c\x3f\x63\x40\x58\x92\x75\x4a\x9d\x24\x10\x1d\x30\x3a\xc2\xf6\x5f\x7d\x79\x85\xcd\x26\xd8\x59\xa4\xeb\xcd\x67\x73\x5b\x07\xee\xda\xd2\x7e\xe9\x4d\x3d\x9a\x8f\xaa\x1e\x21\xe6\x84\x24\x52\xf4\x58\x8c\x7b\xae\x13\x82\x48\x2b\x85\x92\xaa\x59\x2d\x88\xa7\xf9\x0e\x2c\x6c\x6e\x5d\xd3\x9b\x61\x09\xbb\xaa\xc3\xe5\xda\xb9\xa1\xdd\xbf\x0d\xf5\x38\x93\x13\x9e\xc3\xb9\xf7\x5c\x89\x94\xab\x33\x1c\xf1\xec\x73\xee\xee\xe1\x02\x16\xc0\x49\xd0\x21\xd3\x04\x2a\x3b\x89\x38\xfc\x5b\xaa\x96\xe2\x5d\xa3\xbf\xc5\xea\x6e\x97\xf7\xb0\xcd\x32\xf9\xb0\x79\xdf\x6e\xfe\x06\x21\xfb\x0d\x00\x00\xff\xff\xd7\x6e\x59\xfa\x01\x02\x00\x00")

func _003_initUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__003_initUpSql,
		"003_init.up.sql",
	)
}

func _003_initUpSql() (*asset, error) {
	bytes, err := _003_initUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "003_init.up.sql", size: 513, mode: os.FileMode(0644), modTime: time.Unix(1685339920, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x75, 0xd2, 0x36, 0x14, 0xc3, 0x4c, 0xa4, 0xac, 0xf7, 0x90, 0x80, 0xf8, 0x8d, 0x76, 0x1e, 0x5e, 0x33, 0xb5, 0xff, 0x6, 0x7d, 0xc2, 0x63, 0x3f, 0x52, 0xc2, 0x34, 0xe6, 0x6d, 0x78, 0x10, 0xc8}}
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
	"001_init.down.sql":            _001_initDownSql,
	"001_init.up.sql":              _001_initUpSql,
	"002_plural_and_name.down.sql": _002_plural_and_nameDownSql,
	"002_plural_and_name.up.sql":   _002_plural_and_nameUpSql,
	"003_init.down.sql":            _003_initDownSql,
	"003_init.up.sql":              _003_initUpSql,
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
	"001_init.down.sql":            {_001_initDownSql, map[string]*bintree{}},
	"001_init.up.sql":              {_001_initUpSql, map[string]*bintree{}},
	"002_plural_and_name.down.sql": {_002_plural_and_nameDownSql, map[string]*bintree{}},
	"002_plural_and_name.up.sql":   {_002_plural_and_nameUpSql, map[string]*bintree{}},
	"003_init.down.sql":            {_003_initDownSql, map[string]*bintree{}},
	"003_init.up.sql":              {_003_initUpSql, map[string]*bintree{}},
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