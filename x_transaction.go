package twitterscraper

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var (
	indicesRegex      = regexp.MustCompile(`\(\w{1}\[(\d{1,2})\],\s*16\)`)
	onDemandFileRegex = regexp.MustCompile(`['|"]{1}ondemand\.s['|"]{1}:\s*['|"]{1}([\w]*)['|"]{1}`)
)

const (
	additionalRandomNumber = 3
	defaultKeyword         = "obfiowerehiring"
)

// GetTransactionID generates x-client-transaction-id
func (s *Scraper) GetTransactionID(method, path string) (string, error) {
	// Check cache (10 mins)
	if s.txCtx == nil || time.Since(s.txCtx.updatedAt) > 10*time.Minute {
		if err := s.updateTransactionContext(); err != nil {
			return "", err
		}
	}

	timeNow := float64(time.Now().UnixNano()/1e6) / 1000.0
	offset := 1682924400.0
	timeNow = math.Floor(timeNow - offset)
	timeNowInt := int64(timeNow)

	timeNowBytes := make([]byte, 4)
	for i := 0; i < 4; i++ {
		timeNowBytes[i] = byte((timeNowInt >> (i * 8)) & 0xFF)
	}

	hashInput := fmt.Sprintf("%s!%s!%d%s%s", method, path, timeNowInt, defaultKeyword, s.txCtx.animationKey)
	sha := sha256.Sum256([]byte(hashInput))
	hashBytes := sha[:]

	randomNum := byte(rand.Intn(256))

	bytesArr := []byte{}
	for _, k := range s.txCtx.keyBytes {
		bytesArr = append(bytesArr, byte(k))
	}
	bytesArr = append(bytesArr, timeNowBytes...)
	bytesArr = append(bytesArr, hashBytes[:16]...)
	bytesArr = append(bytesArr, byte(additionalRandomNumber))

	out := []byte{randomNum}
	for _, item := range bytesArr {
		out = append(out, item^randomNum)
	}

	return strings.TrimRight(base64.StdEncoding.EncodeToString(out), "="), nil
}

func (s *Scraper) updateTransactionContext() error {
	req, _ := http.NewRequest("GET", "https://x.com", nil)
	req.Header.Set("User-Agent", s.userAgent)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	doc, err := html.Parse(strings.NewReader(bodyString))
	if err != nil {
		return err
	}

	key := findMetaContent(doc, "twitter-site-verification")
	if key == "" {
		return fmt.Errorf("site verification key not found")
	}

	keyBytesDecoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return err
	}
	keyBytes := make([]int, len(keyBytesDecoded))
	for i, b := range keyBytesDecoded {
		keyBytes[i] = int(b)
	}

	frames, err := findFrames(doc, keyBytes)
	if err != nil {
		return err
	}

	match := onDemandFileRegex.FindStringSubmatch(bodyString)
	if len(match) < 2 {
		return fmt.Errorf("ondemand file not found")
	}
	filename := match[1]
	ondemandURL := fmt.Sprintf("https://abs.twimg.com/responsive-web/client-web/ondemand.s.%sa.js", filename)

	reqJS, _ := http.NewRequest("GET", ondemandURL, nil)
	reqJS.Header.Set("User-Agent", s.userAgent)
	respJS, err := s.client.Do(reqJS)
	if err != nil {
		return err
	}
	defer respJS.Body.Close()
	jsBytes, _ := io.ReadAll(respJS.Body)
	jsString := string(jsBytes)

	matches := indicesRegex.FindAllStringSubmatch(jsString, -1)
	if len(matches) == 0 {
		return fmt.Errorf("indices not found")
	}
	indices := []int{}
	for _, m := range matches {
		val, _ := strconv.Atoi(m[1])
		indices = append(indices, val)
	}

	rowIndex := indices[0]
	keyByteIndices := indices[1:]

	rIndex := keyBytes[rowIndex] % 16

	frameTime := 1
	for _, idx := range keyByteIndices {
		frameTime *= (keyBytes[idx] % 16)
	}

	frameTimeFloat := round(float64(frameTime)/10.0) * 10.0

	if rIndex >= len(frames) {
		rIndex = 0
	}
	frameRow := frames[rIndex]
	targetTime := frameTimeFloat / 4096.0

	animationKey := animate(frameRow, targetTime)

	s.txCtx = &transactionContext{
		keyBytes:     keyBytes,
		animationKey: animationKey,
		updatedAt:    time.Now(),
	}
	return nil
}

func findMetaContent(n *html.Node, name string) string {
	if n.Type == html.ElementNode && n.Data == "meta" {
		var nAttr, cAttr string
		for _, a := range n.Attr {
			if a.Key == "name" {
				nAttr = a.Val
			}
			if a.Key == "content" {
				cAttr = a.Val
			}
		}
		if nAttr == name {
			return cAttr
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res := findMetaContent(c, name)
		if res != "" {
			return res
		}
	}
	return ""
}

func findFrames(doc *html.Node, keyBytes []int) ([][]int, error) {
	var frames []*html.Node
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, a := range n.Attr {
				if a.Key == "id" && strings.HasPrefix(a.Val, "loading-x-anim") {
					frames = append(frames, n)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if len(frames) == 0 {
		return nil, fmt.Errorf("no frames found")
	}

	targetFrameIndex := keyBytes[5] % 4
	if targetFrameIndex >= len(frames) {
		targetFrameIndex = 0
	}
	targetFrame := frames[targetFrameIndex]

	// Navigate exactly like Python: list(list(frame.children)[0].children)[1].get("d")
	// Python .children includes text nodes, Go FirstChild/NextSibling also includes text nodes

	// Get children[0] - first child (could be text or element)
	firstChild := targetFrame.FirstChild
	if firstChild == nil {
		return nil, fmt.Errorf("frame has no children")
	}

	// Get children[1] of firstChild - second child
	childIdx := 0
	var pathElement *html.Node
	for c := firstChild.FirstChild; c != nil; c = c.NextSibling {
		if childIdx == 1 {
			pathElement = c
			break
		}
		childIdx++
	}

	if pathElement == nil {
		return nil, fmt.Errorf("could not find path element at children[0].children[1]")
	}

	// Get d attribute
	var dAttr string
	for _, a := range pathElement.Attr {
		if a.Key == "d" {
			dAttr = a.Val
			break
		}
	}

	if dAttr == "" {
		return nil, fmt.Errorf("path d not found")
	}

	if len(dAttr) < 9 {
		return nil, fmt.Errorf("d attr too short")
	}
	dContent := dAttr[9:]
	parts := strings.Split(dContent, "C")

	result := [][]int{}
	re := regexp.MustCompile(`[^\d]+`)

	for _, part := range parts {
		cleaned := re.ReplaceAllString(part, " ")
		numsStr := strings.Fields(cleaned)
		row := []int{}
		for _, ns := range numsStr {
			val, _ := strconv.Atoi(ns)
			row = append(row, val)
		}
		if len(row) > 0 {
			result = append(result, row)
		}
	}

	return result, nil
}

func animate(frames []int, targetTime float64) string {
	fromColor := []float64{float64(frames[0]), float64(frames[1]), float64(frames[2]), 1.0}
	toColor := []float64{float64(frames[3]), float64(frames[4]), float64(frames[5]), 1.0}

	fromRotation := []float64{0.0}
	toRotation := []float64{solve(float64(frames[6]), 60.0, 360.0, true)}

	remainingFrames := frames[7:]
	curves := []float64{}
	for i, item := range remainingFrames {
		curves = append(curves, solve(float64(item), isOdd(float64(i)), 1.0, false))
	}

	c := newCubic(curves)
	val := c.getValue(targetTime)

	color := interpolate(fromColor, toColor, val)
	for i, v := range color {
		if v < 0 {
			color[i] = 0
		}
		if v > 255 {
			color[i] = 255
		}
	}

	rotation := interpolate(fromRotation, toRotation, val)
	matrix := convertRotationToMatrix(rotation[0])

	strArr := []string{}
	for i := 0; i < 3; i++ {
		strArr = append(strArr, fmt.Sprintf("%x", int(round(color[i]))))
	}

	for _, value := range matrix {
		rounded := round2(value)
		if rounded < 0 {
			rounded = -rounded
		}
		hexValue := floatToHex(rounded)
		if strings.HasPrefix(hexValue, ".") {
			hexValue = "0" + hexValue
		} else if hexValue == "" {
			hexValue = "0"
		}
		strArr = append(strArr, strings.ToLower(hexValue))
	}

	strArr = append(strArr, "0", "0")

	joined := strings.Join(strArr, "")
	joined = strings.ReplaceAll(joined, ".", "")
	joined = strings.ReplaceAll(joined, "-", "")

	return joined
}

func solve(value, minVal, maxVal float64, rounding bool) float64 {
	result := value*(maxVal-minVal)/255 + minVal
	if rounding {
		return math.Floor(result)
	}
	return round2(result)
}
