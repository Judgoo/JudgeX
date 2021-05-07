package flake

import (
	"encoding/hex"
	"log"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/osamingo/indigo"
	"github.com/zeebo/blake3"
)

var g *indigo.Generator
var hasher *blake3.Hasher

func init() {
	// 设置随机种子
	t := time.Unix(955641600, 0) // 需要是一个固定的时间
	rand.Seed(int64(time.Since(t)))
	g = indigo.New(nil, indigo.StartTime(t))
	hasher = blake3.New()
	_, err := g.NextID()
	if err != nil {
		log.Fatalln(err)
	}
}

func NextID() string {
	id, err := g.NextID()
	if err != nil {
		id = utils.UUID()
	}
	return id
}

func Digest() string {
	id := NextID()
	hasher.Reset()
	hasher.Write([]byte(id))
	return hex.EncodeToString(hasher.Sum(nil))[:32]
}
