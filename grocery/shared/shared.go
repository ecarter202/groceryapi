package shared

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"

	"grocery/config"

	uuid "github.com/kevinburke/go.uuid"
)

const (
	MODE_DEBUG = "DEBUG"
)

var (
	MODE = ""

	ShutdownChan = make(chan int)
	SigChannel   = make(chan os.Signal, 1)
	_alphaNum    = regexp.MustCompile(`^[a-zA-Z0-9]*$`)
)

func SetDebug() {
	MODE = MODE_DEBUG
	config.APIURL = config.DEVAPIURL
	// set dbhost/dbname if connecting to a dev database
}

func String(item interface{}) string {
	b, err := json.MarshalIndent(item, "", "    ")
	if err != nil {
		return ""
	}

	return "\n" + string(b) + "\n"
}

func HashSha1(item string) string {
	h := sha1.New()
	h.Write([]byte(item))
	hash := hex.EncodeToString(h.Sum(nil))

	return hash
}

func GenProductCode() (c string) {
	u := uuid.NewV4().String()
	x := strings.Split(u, "-")
	c = strings.ToUpper(
		fmt.Sprintf("%s-%s-%s-%s",
			x[0][:4],
			x[1][:4],
			x[2][:4],
			x[3][:4],
		),
	)

	return
}

func IsAlphaNum(s string) bool {
	s = strings.ReplaceAll(s, " ", "")
	return _alphaNum.MatchString(s)
}

func RoundFloat(number float64, places int) float64 {
	f := math.Pow(10, float64(places))
	return math.Round(number*f) / f
}
