package bilibili

import "strings"

// b站的BVID与AVID之间的互相转换，从这段python代码翻译过来的
/*
xorCode = 23442827791579
maskCode = 2251799813685247
maxAid = 1 << 51
alphabet = "FcwAPNKTMug3GV5Lj7EJnHpWsx4tb8haYeviqBz6rkCy12mUSDQX9RdoZf"
encodeMap = 8, 7, 0, 5, 1, 3, 2, 4, 6
decodeMap = tuple(reversed(encodeMap))

bvToAvBase = len(alphabet)
prefix = "BV1"
prefixLen = len(prefix)
codeLen = len(encodeMap)

def av2bv(aid: int) -> str:
    bvid = [""] * 9
    tmp = (maxAid | aid) ^ xorCode
    for i in range(codeLen):
        bvid[encodeMap[i]] = alphabet[tmp % bvToAvBase]
        tmp //= bvToAvBase
    return prefix + "".join(bvid)

def bv2av(bvid: str) -> int:
    assert bvid[:3] == prefix

    bvid = bvid[3:]
    tmp = 0
    for i in range(codeLen):
        idx = alphabet.index(bvid[decodeMap[i]])
        tmp = tmp * bvToAvBase + idx
    return (tmp & maskCode) ^ xorCode

assert av2bv(111298867365120) == "BV1L9Uoa9EUx"
assert bv2av("BV1L9Uoa9EUx") == 111298867365120
*/
var (
	xorCode    = 23442827791579
	maskCode   = 2251799813685247
	maxAid     = 1 << 51
	alphabet   = "FcwAPNKTMug3GV5Lj7EJnHpWsx4tb8haYeviqBz6rkCy12mUSDQX9RdoZf"
	encodeMap  = []int{8, 7, 0, 5, 1, 3, 2, 4, 6}
	decodeMap  = []int{6, 4, 2, 3, 1, 5, 0, 7, 8}
	prefix     = "BV1"
	prefixLen  = len(prefix)
	codeLen    = len(encodeMap)
	bvToAvBase = len(alphabet)
)

func Av2Bv(aid int) string {
	bvid := make([]byte, 9)
	tmp := (maxAid | aid) ^ xorCode
	for i := 0; i < codeLen; i++ {
		bvid[encodeMap[i]] = alphabet[tmp%bvToAvBase]
		tmp /= bvToAvBase
	}
	return prefix + string(bvid)
}

func Bv2Av(bvid string) int {
	if bvid[:prefixLen] != prefix {
		panic("bvid格式错误")
	}
	bvid = bvid[prefixLen:]
	tmp := 0
	for i := 0; i < codeLen; i++ {
		idx := strings.IndexByte(alphabet, bvid[decodeMap[i]])
		tmp = tmp*bvToAvBase + idx
	}
	return (tmp & maskCode) ^ xorCode
}
