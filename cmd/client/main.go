package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ysomad/gigabg/client"
	"github.com/ysomad/gigabg/game/cards"
	"github.com/ysomad/gigabg/ui"
	"github.com/ysomad/gigabg/ui/scene"
	"github.com/ysomad/gigabg/ui/widget"
)

func main() {
	serverAddr := flag.String("addr", "localhost:8080", "game server address")
	flag.Parse()
	cardStore, err := cards.New()
	if err != nil {
		slog.Error("load cards failed", "error", err)
		return
	}

	app, err := ui.NewApp()
	if err != nil {
		slog.Error("init app failed", "error", err)
		return
	}

	httpClient := client.New(*serverAddr)

	connectAndPlay := func(p *widget.Popup, playerID, lobbyID string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		p.SetMessage("Connecting to server...")
		slog.Info("connecting", "player", playerID, "lobby", lobbyID)

		gc, err := client.NewGameClient(ctx, *serverAddr, playerID, lobbyID)
		if err != nil {
			slog.Error("connection failed", "error", err)
			p.SetTitle("Error")
			p.SetMessage(err.Error())
			p.ShowButton("Close", func() { app.HideOverlay() })
			return
		}

		p.SetMessage("Waiting for game state...")
		slog.Info("waiting for state", "player", playerID, "lobby", lobbyID)

		if err := gc.WaitForState(ctx); err != nil {
			if cerr := gc.Close(); cerr != nil {
				slog.Error("close game client", "error", cerr)
			}
			slog.Error("wait for state failed", "error", err)
			p.SetTitle("Error")
			p.SetMessage(err.Error())
			p.ShowButton("Close", func() { app.HideOverlay() })
			return
		}

		slog.Info("joined game", "player", playerID, "lobby", lobbyID)
		app.HideOverlay()
		app.SwitchScene(scene.NewGame(gc, cardStore, app.Font()))
	}

	onJoin := func(playerID, lobbyID string) {
		p := widget.NewPopup(app.Font(), "", "Connecting...")
		app.ShowOverlay(p)

		go func() {
			connectAndPlay(p, playerID, lobbyID)
		}()
	}

	onCreate := func(playerID string, lobbySize int) {
		p := widget.NewPopup(app.Font(), "", "Creating lobby...")
		app.ShowOverlay(p)

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			slog.Info("creating lobby", "player", playerID, "size", lobbySize)

			lobbyID, err := httpClient.CreateLobby(ctx, lobbySize)
			if err != nil {
				slog.Error("create lobby failed", "error", err)
				p.SetTitle("Error")
				p.SetMessage(err.Error())
				p.ShowButton("Close", func() { app.HideOverlay() })
				return
			}

			slog.Info("lobby created", "lobby", lobbyID)
			connectAndPlay(p, playerID, lobbyID)
		}()
	}

	app.SwitchScene(scene.NewMenu(app.Font(), onJoin, onCreate))

	ebiten.SetWindowSize(ui.BaseWidth, ui.BaseHeight)
	ebiten.SetWindowTitle("GIGA Battlegrounds")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(app); err != nil {
		slog.Error("run game failed", "error", err)
		os.Exit(1)
	}
}
