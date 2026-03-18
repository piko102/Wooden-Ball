package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"image/color"
	"log"
	"math"
	"os"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed Micro5-Regular.ttf
var fontFile []byte

//go:embed belka.png
var paddleFile []byte

//go:embed klocek.png
var obstacleFile []byte

//go:embed pilka.png
var ballFile []byte

//go:embed tlo.png
var backgroundFile []byte

var (
	gameFontFace    font.Face
	paddleImage     *ebiten.Image
	obstacleImage   *ebiten.Image
	ballImage       *ebiten.Image
	backgroundImage *ebiten.Image
)

type ScoreEntry struct {
	Name   string `json:"name"`
	Points int    `json:"points"`
}

func init() {
	_ = embed.FS{}

	parsedFont, err := opentype.Parse(fontFile)
	if err != nil {
		log.Fatalf("Error parsing font: %v", err)
	}

	const fontSize = 24.0
	gameFontFace, err = opentype.NewFace(parsedFont, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatalf("Error creating font face: %v", err)
	}

	img, _, err := image.Decode(bytes.NewReader(paddleFile))
	if err != nil {
		log.Fatalf("Error decoding paddle image: %v", err)
	}
	paddleImage = ebiten.NewImageFromImage(img)

	obsImg, _, err := image.Decode(bytes.NewReader(obstacleFile))
	if err != nil {
		log.Fatalf("Error decoding obstacle image: %v", err)
	}
	obstacleImage = ebiten.NewImageFromImage(obsImg)

	ballImg, _, err := image.Decode(bytes.NewReader(ballFile))
	if err != nil {
		log.Fatalf("Error decoding ball image: %v", err)
	}
	ballImage = ebiten.NewImageFromImage(ballImg)

	bgImg, _, err := image.Decode(bytes.NewReader(backgroundFile))
	if err != nil {
		log.Fatalf("Error decoding background image: %v", err)
	}
	backgroundImage = ebiten.NewImageFromImage(bgImg)
}

const (
	screenWidth    = 640
	screenHeight   = 480
	ballRadius     = 10
	paddleWidth    = 23
	paddleHeight   = 80
	obstacleWidth  = 39
	obstacleHeight = 40
	scoreFile      = "scores.json"
)

type Object struct {
	x, y, w, h float32
}

type Game struct {
	ballX, ballY float32
	ballVX, ballVY float32
	paddleY      float32
	obstacles    []Object
	gameOver     bool
	bounceCount  int
	
	highScores   []ScoreEntry
	playerName   string
	enteringName bool
	canRestart   bool

	isScaled     bool
	gameStarted  bool
}

func (g *Game) loadScores() {
	data, err := os.ReadFile(scoreFile)
	if err != nil {
		g.highScores = []ScoreEntry{}
		return
	}
	json.Unmarshal(data, &g.highScores)
}

func (g *Game) saveScores() {
	data, _ := json.Marshal(g.highScores)
	os.WriteFile(scoreFile, data, 0644)
}

func (g *Game) Init() {
	g.ballX = screenWidth / 2
	g.ballY = screenHeight / 2
	g.ballVX = 1.25
	g.ballVY = 1.25
	g.paddleY = screenHeight/2 - paddleHeight/2

	g.obstacles = []Object{
		{screenWidth/2 - obstacleWidth/2, screenHeight/2 - 100, obstacleWidth, obstacleHeight},
		{screenWidth/2 - obstacleWidth/2, screenHeight/2 + 60, obstacleWidth, obstacleHeight},
		{screenWidth/2 - 100, screenHeight/2 - obstacleHeight/2, obstacleWidth, obstacleHeight},
	}
	g.bounceCount = 0
	g.gameOver = false
	g.enteringName = false
	g.playerName = ""
	g.canRestart = false
	g.gameStarted = false
	g.loadScores()
}

func (g *Game) Update() error {
	if !g.gameOver && inpututil.IsKeyJustPressed(ebiten.KeyR) {
		if g.isScaled {
			ebiten.SetWindowSize(screenWidth, screenHeight)
			g.isScaled = false
		} else {
			ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
			g.isScaled = true
		}
	}

	if !g.gameStarted {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.gameStarted = true
		}
		return nil
	}

	if g.gameOver {
		if g.enteringName {
			chars := ebiten.AppendInputChars(nil)
			if len(g.playerName) < 10 {
				g.playerName += string(chars)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(g.playerName) > 0 {
				g.playerName = g.playerName[:len(g.playerName)-1]
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && len(g.playerName) > 0 {
				g.highScores = append(g.highScores, ScoreEntry{Name: g.playerName, Points: g.bounceCount})
				sort.Slice(g.highScores, func(i, j int) bool {
					return g.highScores[i].Points > g.highScores[j].Points
				})
				if len(g.highScores) > 10 {
					g.highScores = g.highScores[:10]
				}
				g.saveScores()
				g.enteringName = false
				g.canRestart = true
			}
			return nil
		}
		if g.canRestart {
			if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyEnter) {
				g.Init()
				g.gameStarted = true
			}
		} else {
			if len(g.highScores) < 10 || g.bounceCount > g.highScores[len(g.highScores)-1].Points {
				g.enteringName = true
			} else {
				g.canRestart = true
			}
		}
		return nil
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) { g.paddleY -= 3 }
	if ebiten.IsKeyPressed(ebiten.KeyDown) { g.paddleY += 3 }
	if g.paddleY < 0 { g.paddleY = 0 }
	if g.paddleY > screenHeight-paddleHeight { g.paddleY = screenHeight - paddleHeight }

	g.ballX += g.ballVX
	g.ballY += g.ballVY

	if g.ballY-ballRadius < 0 {
		g.ballY = ballRadius
		g.ballVY = -g.ballVY
		g.bounceCount++
	} else if g.ballY+ballRadius > screenHeight {
		g.ballY = screenHeight - ballRadius
		g.ballVY = -g.ballVY
		g.bounceCount++
	}

	if g.ballX-ballRadius < 0 {
		g.ballX = ballRadius
		g.ballVX = -g.ballVX
		g.bounceCount++
	}

	// Poprawiona kolizja z paletką (obsługa krawędzi i kątów)
	px := float32(screenWidth - paddleWidth - 10)
	if g.ballX+ballRadius > px && g.ballX-ballRadius < px+paddleWidth &&
		g.ballY+ballRadius > g.paddleY && g.ballY-ballRadius < g.paddleY+paddleHeight {
		
		dx := g.ballX - (px + paddleWidth/2)
		dy := g.ballY - (g.paddleY + paddleHeight/2)
		overlapX := (ballRadius + paddleWidth/2) - float32(math.Abs(float64(dx)))
		overlapY := (ballRadius + paddleHeight/2) - float32(math.Abs(float64(dy)))

		if overlapX < overlapY {
			// Kolizja boczna (najważniejsza - front paletki)
			if dx < 0 { // Uderzenie w przód paletki
				g.ballX -= overlapX
				g.ballVX = -float32(math.Abs(float64(g.ballVX))) * 1.05
				
				// Steering: zmiana kąta Y zależnie od miejsca uderzenia
				hitPos := (g.ballY - (g.paddleY + paddleHeight/2)) / (paddleHeight / 2)
				g.ballVY = hitPos * 2.0 // Piłka leci pod kątem
			} else { // Uderzenie w tył (rzadkie)
				g.ballX += overlapX
				g.ballVX = float32(math.Abs(float64(g.ballVX)))
			}
		} else {
			// Uderzenie w górny/dolny kant (teraz odbija się naturalnie pionowo)
			if dy > 0 { g.ballY += overlapY } else { g.ballY -= overlapY }
			g.ballVY = -g.ballVY
		}
		g.bounceCount++
	}

	for _, obs := range g.obstacles {
		if g.ballX+ballRadius > obs.x && g.ballX-ballRadius < obs.x+obs.w &&
			g.ballY+ballRadius > obs.y && g.ballY-ballRadius < obs.y+obs.h {
			
			obsCenterX := obs.x + obs.w/2
			obsCenterY := obs.y + obs.h/2
			dx := g.ballX - obsCenterX
			dy := g.ballY - obsCenterY
			overlapX := (ballRadius + obs.w/2) - float32(math.Abs(float64(dx)))
			overlapY := (ballRadius + obs.h/2) - float32(math.Abs(float64(dy)))

			if overlapX < overlapY {
				if dx > 0 { g.ballX += overlapX } else { g.ballX -= overlapX }
				g.ballVX = -g.ballVX
			} else {
				if dy > 0 { g.ballY += overlapY } else { g.ballY -= overlapY }
				g.ballVY = -g.ballVY
			}
			g.bounceCount++
			break 
		}
	}

	if g.ballX > screenWidth {
		g.gameOver = true
		g.loadScores()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(backgroundImage, nil)

	if !g.gameStarted {
		msg := "To play, tap Enter"
		bounds := text.BoundString(gameFontFace, msg)
		text.Draw(screen, msg, gameFontFace, screenWidth/2 - bounds.Dx()/2, screenHeight/2, color.White)
		return
	}

	if !g.gameOver {
		countMsg := fmt.Sprintf("Bounces: %d", g.bounceCount)
		text.Draw(screen, countMsg, gameFontFace, 10, 20, color.White)
	}

	bop := &ebiten.DrawImageOptions{}
	bw, bh := ballImage.Bounds().Dx(), ballImage.Bounds().Dy()
	scaleY := (2.0 * ballRadius) / float64(bh)
	scaleX := scaleY * (458.0 / 438.0)
	bop.GeoM.Scale(scaleX, scaleY)
	bop.GeoM.Translate(float64(g.ballX)-float64(bw)*scaleX/2, float64(g.ballY)-float64(bh)*scaleY/2)
	screen.DrawImage(ballImage, bop)

	paddleX := float32(screenWidth - paddleWidth - 10)
	pop := &ebiten.DrawImageOptions{}
	pw, ph := paddleImage.Bounds().Dx(), paddleImage.Bounds().Dy()
	pop.GeoM.Scale(float64(paddleWidth)/float64(pw), float64(paddleHeight)/float64(ph))
	pop.GeoM.Translate(float64(paddleX), float64(g.paddleY))
	screen.DrawImage(paddleImage, pop)

	ow, oh := obstacleImage.Bounds().Dx(), obstacleImage.Bounds().Dy()
	for _, obs := range g.obstacles {
		oop := &ebiten.DrawImageOptions{}
		oop.GeoM.Scale(float64(obs.w)/float64(ow), float64(obs.h)/float64(oh))
		oop.GeoM.Translate(float64(obs.x), float64(obs.y))
		screen.DrawImage(obstacleImage, oop)
	}

	if g.gameOver {
		screen.Fill(color.Black)
		title := "TOP 10 HIGH SCORES"
		tBounds := text.BoundString(gameFontFace, title)
		text.Draw(screen, title, gameFontFace, screenWidth/2 - tBounds.Dx()/2, 40, color.RGBA{0, 200, 255, 255})
		
		for i, s := range g.highScores {
			entry := fmt.Sprintf("%d. %-12s %d", i+1, s.Name, s.Points)
			text.Draw(screen, entry, gameFontFace, 200, 80+i*22, color.White)
		}

		if g.enteringName {
			msg := "NEW RECORD! ENTER NICKNAME:"
			bounds := text.BoundString(gameFontFace, msg)
			text.Draw(screen, msg, gameFontFace, screenWidth/2 - bounds.Dx()/2, 340, color.RGBA{255, 200, 0, 255})
			
			nameBounds := text.BoundString(gameFontFace, g.playerName+"_")
			text.Draw(screen, g.playerName+"_", gameFontFace, screenWidth/2 - nameBounds.Dx()/2, 380, color.White)
			
			hint := "Press ENTER to save"
			hBounds := text.BoundString(gameFontFace, hint)
			text.Draw(screen, hint, gameFontFace, screenWidth/2 - hBounds.Dx()/2, 420, color.RGBA{150, 150, 150, 255})
		} else {
			restartMsg := "Play again? Press ENTER"
			rBounds := text.BoundString(gameFontFace, restartMsg)
			text.Draw(screen, restartMsg, gameFontFace, screenWidth/2 - rBounds.Dx()/2, 380, color.RGBA{255, 200, 0, 255})
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := &Game{}
	game.isScaled = true
	game.Init()

	ebiten.SetWindowTitle("Simple Pong with Obstacles")
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(120)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
