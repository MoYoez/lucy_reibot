// Package fortune
package fortune

import (
	"encoding/json"
	"fmt"
	"hash/crc64"
	"image"
	"image/color"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/MoYoez/Lucy_reibot/utils/toolchain"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/imgfactory"
	"github.com/fogleman/gg"
	tgba "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/MoYoez/Lucy_reibot/utils/transform"
	ctrl "github.com/FloatTech/zbpctrl"
	rei "github.com/fumiama/ReiBot"
)

type card struct {
	Name string `json:"name"`
	Info struct {
		Description        string `json:"description"`
		ReverseDescription string `json:"reverseDescription"`
		ImgURL             string `json:"imgUrl"`
	} `json:"info"`
}

type cardset = map[string]card

var (
	info     string
	cardMap  = make(cardset, 256)
	position = []string{"正位", "逆位"}
	result   map[int64]int
	signTF   map[string]int
)

func init() {
	engine := rei.Register("fortune", &ctrl.Options[*rei.Ctx]{
		DisableOnDefault:  false,
		Help:              "Hi NekoPachi!\n说明书: https://lucy-sider.lemonkoi.one",
		PrivateDataFolder: "fortune",
	})
	signTF = make(map[string]int)
	result = make(map[int64]int)
	// onload fortune mapset.
	data, err := os.ReadFile(transform.ReturnLucyMainDataIndex("funwork") + "tarots.json")
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &cardMap)
	picDir, err := os.ReadDir(transform.ReturnLucyMainDataIndex("funwork") + "randpic")
	if err != nil {
		return
	}
	picDirNum := len(picDir)
	reg := regexp.MustCompile(`[^.]+`)
	engine.OnMessageCommand("fortune").SetBlock(true).Handle(func(ctx *rei.Ctx) {
		getUserID, getUserName := toolchain.GetChatUserInfoID(ctx)
		userPic := strconv.FormatInt(getUserID, 10) + time.Now().Format("20060102") + ".png"
		usersRandPic := RandSenderPerDayN(getUserID, picDirNum)
		picDirName := picDir[usersRandPic].Name()
		list := reg.FindAllString(picDirName, -1)
		p := rand.Intn(2)
		is := rand.Intn(77)
		i := is + 1
		card := cardMap[(strconv.Itoa(i))]
		if p == 0 {
			info = card.Info.Description
		} else {
			info = card.Info.ReverseDescription
		}
		userS := strconv.FormatInt(getUserID, 10)
		now := time.Now().Format("20060102")
		// modify this possibility to 40-100, don't be to low.
		randEveryone := RandSenderPerDayN(getUserID, 50)
		var si = now + userS // use map to store.
		loadNotoSans := transform.ReturnLucyMainDataIndex("funwork") + "NotoSansCJKsc-Regular.otf"
		if signTF[si] == 0 {
			result[getUserID] = randEveryone + 50
			// background
			img, err := gg.LoadImage(transform.ReturnLucyMainDataIndex("funwork") + "randpic" + "/" + list[0] + ".png")
			if err != nil {
				return
			}
			bgFormat := imgfactory.Limit(img, 1920, 1080)
			getBackGroundMainColorR, getBackGroundMainColorG, getBackGroundMainColorB := GetAverageColorAndMakeAdjust(bgFormat)
			mainContext := gg.NewContext(bgFormat.Bounds().Dx(), bgFormat.Bounds().Dy())
			mainContextWidth := mainContext.Width()
			mainContextHight := mainContext.Height()
			mainContext.DrawImage(bgFormat, 0, 0)
			// draw Round rectangle
			mainContext.SetFontFace(LoadFontFace(loadNotoSans, 50))
			if err != nil {
				_, _ = ctx.SendPlainMessage(false, "Something wrong while rendering pic? font")
				return
			}
			// shade mode || not bugs(
			mainContext.SetLineWidth(4)
			mainContext.SetRGBA255(255, 255, 255, 255)
			mainContext.DrawRoundedRectangle(0, float64(mainContextHight-150), float64(mainContextWidth), 150, 16)
			mainContext.Stroke()
			mainContext.SetRGBA255(255, 224, 216, 215)
			mainContext.DrawRoundedRectangle(0, float64(mainContextHight-150), float64(mainContextWidth), 150, 16)
			mainContext.Fill()
			// avatar,name,desc
			// draw third round rectangle
			mainContext.SetRGBA255(91, 57, 83, 255)
			mainContext.SetFontFace(LoadFontFace(loadNotoSans, 25))
			charCount := 0.0
			setBreaker := false
			emojiRegex := regexp.MustCompile(`[\x{1F600}-\x{1F64F}|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F700}-\x{1F77F}]|[\x{1F780}-\x{1F7FF}]|[\x{1F800}-\x{1F8FF}]|[\x{1F900}-\x{1F9FF}]|[\x{1FA00}-\x{1FA6F}]|[\x{1FA70}-\x{1FAFF}]|[\x{1FB00}-\x{1FBFF}]|[\x{1F170}-\x{1F251}]|[\x{1F300}-\x{1F5FF}]|[\x{1F600}-\x{1F64F}]|[\x{1FC00}-\x{1FCFF}]|[\x{1F004}-\x{1F0CF}]|[\x{1F170}-\x{1F251}]]+`)
			getUserName = emojiRegex.ReplaceAllString(getUserName, "")
			var truncated string
			var UserFloatNum float64
			// set rune count
			for _, runeValue := range getUserName {
				charWidth := utf8.RuneLen(runeValue)
				if charWidth == 3 {
					UserFloatNum = 1.5
				} else {
					UserFloatNum = float64(charWidth)
				}
				if charCount+UserFloatNum > 24 {
					setBreaker = true
					break
				}
				truncated += string(runeValue)
				charCount += UserFloatNum
			}
			if setBreaker {
				getUserName = truncated + "..."
			} else {
				getUserName = truncated
			}
			nameLength, _ := mainContext.MeasureString(getUserName)
			var renderLength float64
			renderLength = nameLength + 160
			if nameLength+160 <= 450 {
				renderLength = 450
			}
			mainContext.DrawRoundedRectangle(50, float64(mainContextHight-175), renderLength, 250, 20)
			mainContext.Fill()
			// avatar draw end.
			avatarFormatRaw := toolchain.GetTargetAvatar(ctx)
			if avatarFormatRaw != nil {
				mainContext.DrawImage(imgfactory.Size(avatarFormatRaw, 100, 100).Circle(0).Image(), 60, int(float64(mainContextHight-150)+25))
			}
			mainContext.SetRGBA255(255, 255, 255, 255)
			mainContext.DrawString("User Info", 60, float64(mainContextHight-150)+10) // basic ui
			mainContext.SetRGBA255(155, 121, 147, 255)
			mainContext.DrawString(getUserName, 180, float64(mainContextHight-150)+50)
			mainContext.DrawString(fmt.Sprintf("今日人品值: %d", randEveryone+50), 180, float64(mainContextHight-150)+100)
			mainContext.Fill()
			// AOSP time and date
			setInlineColor := color.NRGBA{R: uint8(getBackGroundMainColorR), G: uint8(getBackGroundMainColorG), B: uint8(getBackGroundMainColorB), A: 255}
			if err != nil {
				_, _ = ctx.SendPlainMessage(false, "Something wrong while rendering pic?")
				return
			}
			formatTimeDate := time.Now().Format("2006 / 01 / 02")
			formatTimeCurrent := time.Now().Format("15 : 04 : 05")
			formatTimeWeek := time.Now().Weekday().String()
			mainContext.SetFontFace(LoadFontFace(loadNotoSans, 35))
			setOutlineColor := color.White
			DrawBorderString(mainContext, formatTimeCurrent, 5, float64(mainContextWidth-80), 50, 1, 0.5, setInlineColor, setOutlineColor)
			DrawBorderString(mainContext, formatTimeDate, 5, float64(mainContextWidth-80), 100, 1, 0.5, setInlineColor, setOutlineColor)
			DrawBorderString(mainContext, formatTimeWeek, 5, float64(mainContextWidth-80), 150, 1, 0.5, setInlineColor, setOutlineColor)
			mainContext.FillPreserve()
			if err != nil {
				return
			}
			mainContext.SetFontFace(LoadFontFace(loadNotoSans, 140))
			DrawBorderString(mainContext, "|", 5, float64(mainContextWidth-30), 65, 1, 0.5, setInlineColor, setOutlineColor)
			// throw tarot card
			mainContext.SetFontFace(LoadFontFace(loadNotoSans, 20))
			if err != nil {
				_, _ = ctx.SendPlainMessage(false, "Something wrong while rendering pic?")
				return
			}
			mainContext.SetRGBA255(91, 57, 83, 255)
			mainContext.DrawRoundedRectangle(float64(mainContextWidth-300), float64(mainContextHight-350), 450, 300, 20)
			mainContext.Fill()
			mainContext.SetRGBA255(255, 255, 255, 255)
			mainContext.SetLineWidth(3)
			mainContext.DrawString("今日塔罗牌", float64(mainContextWidth-300)+10, float64(mainContextHight-350)+30)
			mainContext.SetRGBA255(155, 121, 147, 255)
			mainContext.DrawString(card.Name, float64(mainContextWidth-300)+10, float64(mainContextHight-350)+60)
			mainContext.DrawString(fmt.Sprintf("- %s", position[p]), float64(mainContextWidth-300)+10, float64(mainContextHight-350)+280)
			placedList := SplitChineseString(info, 44)
			for ist, v := range placedList {
				mainContext.DrawString(v, float64(mainContextWidth-300)+10, float64(mainContextHight-350)+90+float64(ist*30))
			}
			// output
			mainContext.SetFontFace(LoadFontFace(loadNotoSans, 20))
			mainContext.SetRGBA255(186, 163, 157, 255)
			mainContext.DrawStringAnchored("Generated By Lucy (HiMoYo), Design By MoeMagicMango", float64(mainContextWidth-15), float64(mainContextHight-30), 1, 1)
			mainContext.Fill()
			_ = mainContext.SavePNG(engine.DataFolder() + "jrrp/" + userPic)
			_, _ = ctx.SendPhoto(tgba.FilePath(engine.DataFolder()+"jrrp/"+userPic), true, "")
			signTF[si] = 1
		} else {
			_, _ = ctx.SendPhoto(tgba.FilePath(engine.DataFolder()+"jrrp/"+userPic), true, "今天已经测试过了哦w")
		}
	})
}

// SplitChineseString Split Chinese type chart.
func SplitChineseString(s string, length int) []string {
	results := make([]string, 0)
	runes := []rune(s)
	start := 0
	for i := 0; i < len(runes); i++ {
		size := utf8.RuneLen(runes[i])
		if start+size > length {
			results = append(results, string(runes[0:i]))
			runes = runes[i:]
			i, start = 0, 0
		}
		start += size
	}
	if len(runes) > 0 {
		results = append(results, string(runes))
	}
	return results
}

// LoadFontFace load font face once before running, to work it quickly and save memory.
func LoadFontFace(filePath string, size float64) font.Face {
	fontFile, _ := os.ReadFile(filePath)
	fontFileParse, _ := opentype.Parse(fontFile)
	fontFace, _ := opentype.NewFace(fontFileParse, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	return fontFace
}

// GetAverageColorAndMakeAdjust different from k-means algorithm,it uses origin plugin's algorithm.(Reduce the cost of averge color usage.)
func GetAverageColorAndMakeAdjust(image image.Image) (int, int, int) {
	var RList []int
	var GList []int
	var BList []int
	width, height := image.Bounds().Size().X, image.Bounds().Size().Y
	// use the center of the bg, to make it more quickly and save memory and usage.
	for x := int(math.Round(float64(width) / 1.5)); x < int(math.Round(float64(width))); x++ {
		for y := height / 10; y < height/2; y++ {
			r, g, b, _ := image.At(x, y).RGBA()
			RList = append(RList, int(r>>8))
			GList = append(GList, int(g>>8))
			BList = append(BList, int(b>>8))
		}
	}
	RAverage := int(Average(RList))
	GAverage := int(Average(GList))
	BAverage := int(Average(BList))
	return RAverage, GAverage, BAverage
}

// Average sum all the numbers and divide by the length of the list.
func Average(numbers []int) float64 {
	var sum float64
	for _, num := range numbers {
		sum += float64(num)
	}
	return math.Round(sum / float64(len(numbers)))
}

// DrawBorderString GG Package Not support The string render, so I write this (^^)
func DrawBorderString(page *gg.Context, s string, size int, x float64, y float64, ax float64, ay float64, inlineRGB color.Color, outlineRGB color.Color) {
	page.SetColor(outlineRGB)
	n := size
	for dy := -n; dy <= n; dy++ {
		for dx := -n; dx <= n; dx++ {
			if dx*dx+dy*dy >= n*n {
				continue
			}
			renderX := x + float64(dx)
			renderY := y + float64(dy)
			page.DrawStringAnchored(s, renderX, renderY, ax, ay)
		}
	}
	page.SetColor(inlineRGB)
	page.DrawStringAnchored(s, x, y, ax, ay)
}

// RandSenderPerDayN 每个用户每天随机数
func RandSenderPerDayN(uid int64, n int) int {
	sum := crc64.New(crc64.MakeTable(crc64.ISO))
	_, _ = sum.Write(binary.StringToBytes(time.Now().Format("20060102")))
	_, _ = sum.Write((*[8]byte)(unsafe.Pointer(&uid))[:])
	r := rand.New(rand.NewSource(int64(sum.Sum64())))
	return r.Intn(n)
}
