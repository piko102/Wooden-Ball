# Project Overview: Simple Pong with Obstacles

A 2D Pong-style game built using **Go 1.24.0** and the **Ebitengine (ebiten) v2** library. The game features a single-player experience where the player controls a right-side paddle to bounce a ball, avoiding obstacles in the center of the screen.

## Core Technologies
- **Language:** Go 1.24.0
- **Graphics Library:** Ebitengine v2 (`github.com/hajimehoshi/ebiten/v2`)
- **Assets:** Embedded using `go:embed` (PNGs, TTF font)
- **Data Storage:** JSON for high scores (`scores.json`)

## Architecture & Key Components
- **Game Engine:** Operates at **120 TPS** (Ticks Per Second) for high precision.
- **Physics:** 
  - AABB collision resolution with snapping to prevent overlapping.
  - **Steering:** The bounce angle depends on where the ball hits the paddle (top/bottom hits result in steeper angles).
  - Ball speed increases by 5% after each paddle hit.
- **Resolution:**
  - **Logical Resolution:** 640x480 pixels.
  - **Window Scaling:** Default 1280x960 (2x scale).
  - **Toggle:** Press **'R'** to switch between 1x (640x480) and 2x scaling.
- **High Scores:** Tracks Top 10 scores with player nicknames (max 10 chars).

## Building and Running
### Prerequisites
- Go 1.24.0 or higher.
- C compiler (GCC/Clang) for Ebitengine's dependencies.

### Commands
- **Run the game:**
  ```bash
  go run main.go
  ```
- **Build the executable:**
  ```bash
  go build -o pong-game main.go
  ```

## Development Conventions
- **Asset Management:** All binary assets (images, fonts) must be embedded using the `//go:embed` directive and decoded in the `init()` function.
- **Coordination System:** 
  - (0,0) is top-left.
  - Logical coordinates are used for all physics; `ebiten.DrawImageOptions` handles scaling for rendering.
- **Physics Logic:**
  - Update movement and check collisions in the `Update()` method.
  - Ensure `overlapX` and `overlapY` calculations are used for stable collision resolution.
- **Text Rendering:** Use `golang.org/x/image/font` and `opentype` for high-quality font rendering. Center text using `text.BoundString`.
