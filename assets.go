package urknall

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

var assets = map[string][]byte{}

func assetNames() (names []string) {
	for name, _ := range assets {
		names = append(names, name)
	}
	return names
}

var (
	devAssetsPath string
	Development bool
	Debug bool
)

func init() {
	devAssetsPath = os.Getenv("DEV_ASSETS_PATH")
	if devAssetsPath != "" {
		Development = true
	}
}

func logDebug(format string, args ...interface{}) {
	if Debug {
		fmt.Printf(format + "\n", args...)
	}
}

func mustReadAsset(key string) []byte {
	content, e := readAsset(key)
	if e != nil {
		panic(e)
	}
	return content
}

func readAsset(key string) ([]byte, error) {
	if devAssetsPath != "" {
		path := devAssetsPath + "/" + key
		logDebug("reading file from dev path %s", path)
		b, e := ioutil.ReadFile(path)
		if e == nil {
			return b, nil
		}
	}
	b, ok := assets[key]
	if !ok {
		return nil, fmt.Errorf("asset %s not found in %v", key, assetNames())
	}
	gz, err := gzip.NewReader(bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("Decompression failed: %s", err.Error())
	}

	var buf bytes.Buffer
	io.Copy(&buf, gz)
	gz.Close()
	return buf.Bytes(), nil
}

func init() {
	assets["fw_ipset.conf"] = []byte{
0x1f,0x8b,0x8,0x0,0x0,0x9,0x6e,0x88,0x0,0xff,0xaa,0xae,
0x56,0x28,0x4a,0xcc,0x4b,0x4f,0x55,0xd0,0xf3,0xc,0x8,0x4e,
0x2d,0x29,0x56,0xa8,0xad,0x5,0xa,0x41,0x38,0x41,0xa9,0xc5,
0x25,0xf9,0x45,0xa9,0x10,0xa1,0xd4,0xbc,0x14,0x20,0x83,0xb,
0x0,0x0,0x0,0xff,0xff,0x1,0x0,0x0,0xff,0xff,0x59,0x6f,
0x68,0xbd,0x30,0x0,0x0,0x0,

	}
	assets["fw_rules_ipv4.conf"] = []byte{
0x1f,0x8b,0x8,0x0,0x0,0x9,0x6e,0x88,0x0,0xff,0xa4,0x54,
0x41,0x4f,0xe3,0x3c,0x14,0xbc,0xf7,0x57,0x3c,0x89,0xb,0x20,
0x12,0xb5,0xc0,0xc7,0xc7,0xf6,0x16,0x9a,0xb0,0x8d,0x4,0x49,
0xd4,0x84,0xe5,0xb0,0xda,0x83,0x71,0x5c,0xea,0x6d,0x6a,0x67,
0x1d,0x47,0xa8,0x42,0xfd,0xef,0xfb,0xdc,0x26,0xa1,0x29,0x84,
0xa2,0xee,0xc9,0x71,0xec,0x37,0x6f,0x3c,0x33,0xf6,0xe9,0x94,
0x67,0x9a,0xa9,0xde,0xd0,0xf,0xa2,0x87,0x4,0xdc,0x49,0x18,
0xc1,0xcf,0xfe,0xb0,0xff,0xab,0x37,0xbc,0xd,0x27,0x8f,0xce,
0xc4,0x6d,0xfd,0xb,0x1f,0x92,0x9d,0x6d,0xbd,0x23,0x70,0x28,
0x65,0xb9,0x6,0x22,0x96,0xa0,0x58,0x46,0x34,0x4b,0x41,0x2a,
0x60,0x85,0x26,0x4f,0x19,0x2f,0x66,0x38,0xa5,0x52,0x8,0x46,
0x35,0x97,0xa2,0xb0,0x7b,0x96,0xf,0x9b,0x66,0x3,0xb0,0x16,
0x80,0xbb,0x34,0x3,0xcb,0xda,0x8c,0x13,0xef,0xce,0x49,0x3c,
0xf7,0xcc,0x8b,0x13,0xe7,0xe6,0xce,0x8f,0xc7,0x9e,0xb,0xd6,
0x6f,0x70,0x46,0x23,0x2f,0x4a,0x4c,0x65,0x4d,0xea,0x90,0xda,
0x8a,0xfc,0x7,0xa5,0xaf,0xaf,0xc0,0xa7,0x20,0xa4,0x6,0x3b,
0x22,0x8a,0x8,0xc9,0x53,0x58,0xad,0x2,0xef,0xf1,0xc,0x57,
0x98,0x30,0x93,0xcf,0xd1,0x8d,0xa,0x59,0x26,0x5f,0x80,0x64,
0x19,0x68,0x45,0xa6,0x53,0x4e,0x41,0xa,0xd0,0x33,0x6,0x99,
0x94,0xf9,0x13,0xa1,0x73,0xe0,0x2,0x95,0x9e,0x12,0xca,0x50,
0x3,0xa7,0xd2,0xc0,0xe2,0xb8,0xbe,0x4d,0xd3,0xa9,0x69,0x5a,
0xb2,0xbd,0xd2,0xdb,0x90,0xb4,0x1f,0xb9,0x9e,0xfd,0x88,0x2,
0xa4,0xf4,0x59,0x53,0xb3,0xa3,0xa3,0x9f,0x2e,0x45,0xbf,0xab,
0xe3,0xce,0x5a,0x73,0xfa,0xa6,0xfb,0x96,0x3c,0xd8,0x3e,0x2c,
0xf5,0x93,0x2c,0x71,0x87,0x1b,0xc4,0xe6,0x9c,0xf3,0x32,0x2f,
0xda,0x80,0x58,0x66,0xfb,0x35,0xf,0x2c,0x2,0x2b,0x87,0x32,
0xcd,0x8d,0x5,0xeb,0xc1,0x4a,0x73,0xa9,0x34,0xfc,0x77,0xd1,
0x16,0xb3,0x1,0x8e,0xfc,0xe0,0x3b,0xa6,0xea,0x4f,0x89,0x71,
0x6a,0x41,0xe7,0xc0,0xe9,0x22,0xef,0xa8,0xa,0x98,0x7e,0x91,
0x6a,0xe,0x9,0x5f,0x30,0x88,0x94,0xd4,0x92,0xca,0xc,0x8e,
0x83,0x24,0x3a,0xa9,0xc1,0xda,0x58,0xdb,0x5c,0x6,0xe7,0x17,
0x26,0x18,0x6f,0xdf,0xef,0x5d,0x96,0xcd,0xb9,0xc7,0xa3,0xa8,
0x46,0x4,0xb,0x62,0x89,0xfd,0x66,0x12,0xa9,0xc2,0xf1,0x1d,
0x17,0x32,0x65,0x27,0x40,0x4a,0x2d,0x17,0x44,0x73,0x8a,0x2e,
0x2d,0x81,0x14,0x5,0x7f,0xde,0x58,0x94,0x2b,0xbe,0x20,0x6a,
0x9,0x7e,0xd4,0xcd,0xe5,0xea,0xff,0xe1,0xd5,0x75,0xc3,0xa6,
0x9a,0xed,0xa6,0x17,0x73,0xda,0xa1,0xc3,0x38,0x49,0xa2,0xfd,
0x7e,0x68,0xba,0xf6,0x63,0x3d,0x54,0x7d,0xaf,0xfb,0xfb,0xda,
0x1c,0x82,0x7a,0x79,0x79,0xb1,0xf,0x76,0x2b,0x71,0x47,0x10,
0xc7,0xe3,0x56,0x72,0xbf,0xd8,0xe6,0xfc,0x7c,0xaf,0x46,0x88,
0x84,0x39,0x7e,0x66,0x60,0xdf,0x72,0xc5,0x5e,0xcc,0x5,0x5a,
0xad,0xc,0xfc,0xed,0xfa,0x39,0xdc,0x4c,0x76,0xa2,0xef,0x17,
0xae,0xa4,0x73,0xa6,0x6e,0x4a,0x9e,0xa5,0x63,0x34,0xd9,0x2c,
0x22,0xbd,0xfa,0x49,0x42,0x82,0xe9,0x7a,0xc3,0x5e,0xed,0xde,
0x63,0xb7,0x2e,0x75,0x63,0x5f,0x98,0x33,0x61,0xfe,0xd6,0x77,
0xfb,0xd8,0x24,0xd,0xf9,0xa6,0xa0,0x65,0xfd,0xa6,0x9a,0xcf,
0xea,0xc2,0x9f,0xd8,0x7,0xb9,0x32,0x18,0x7c,0xbb,0xfc,0x67,
0xb7,0x3f,0xb8,0xd3,0x5f,0xc1,0x7d,0x53,0x62,0x14,0xde,0xdf,
0xfb,0xe8,0xcc,0xa9,0x20,0xba,0xdb,0x9e,0xc0,0x49,0x5a,0xde,
0x54,0x55,0x7f,0x1,0x0,0x0,0xff,0xff,0x1,0x0,0x0,0xff,
0xff,0xec,0xed,0xe,0xf2,0xc6,0x6,0x0,0x0,

	}
	assets["fw_rules_ipv6.conf"] = []byte{
0x1f,0x8b,0x8,0x0,0x0,0x9,0x6e,0x88,0x0,0xff,0xd2,0x4a,
0xcb,0xcc,0x29,0x49,0x2d,0xe2,0xb2,0xf2,0xf4,0xb,0x8,0xd,
0x51,0x70,0x9,0xf2,0xf,0x50,0x88,0x36,0xb0,0x32,0x88,0xe5,
0xb2,0x72,0xf3,0xf,0xa,0x77,0xc,0x72,0x41,0x11,0xf3,0xf,
0xd,0x41,0x53,0xc6,0xe5,0xec,0xef,0xeb,0xeb,0x19,0xc2,0x5,
0x0,0x0,0x0,0xff,0xff,0x1,0x0,0x0,0xff,0xff,0xaa,0xef,
0x95,0xad,0x49,0x0,0x0,0x0,

	}
	assets["fw_upstart.sh"] = []byte{
0x1f,0x8b,0x8,0x0,0x0,0x9,0x6e,0x88,0x0,0xff,0x8c,0xcb,
0xb1,0xa,0xc2,0x30,0x14,0x85,0xe1,0x39,0xf7,0x29,0x6e,0xab,
0x8b,0x43,0xcc,0x22,0x5d,0xea,0x22,0xa2,0xd0,0xa7,0x90,0x34,
0x1c,0x31,0x20,0xb1,0xe4,0xa6,0x2e,0xa5,0xef,0x6e,0x42,0x15,
0xc4,0xc9,0xe5,0xc,0x87,0xff,0x5b,0x55,0xa6,0xf7,0xc1,0xc8,
0x8d,0x4,0x89,0x35,0x88,0x9c,0x15,0x70,0xbd,0xee,0xce,0x87,
0xe3,0xa9,0x66,0x1f,0x48,0x4d,0x13,0x6f,0xbb,0x90,0x10,0xaf,
0xd6,0x81,0xe7,0x79,0x43,0x4a,0x99,0x51,0xa2,0x91,0x42,0xfd,
0x50,0x64,0x84,0xa4,0x47,0x4,0xeb,0x8a,0xf7,0x6c,0x90,0x5c,
0xfe,0x93,0xed,0xef,0x90,0x25,0x90,0x62,0xde,0xfd,0xf2,0xeb,
0xf,0xf9,0xed,0xe3,0x98,0xf7,0xe2,0x87,0xe7,0xee,0xcb,0x34,
0xff,0xa2,0x26,0xa3,0xb6,0x25,0x88,0x75,0x44,0x2f,0x0,0x0,
0x0,0xff,0xff,0x1,0x0,0x0,0xff,0xff,0xbf,0x86,0x2c,0xd6,
0xde,0x0,0x0,0x0,

	}
	
}
