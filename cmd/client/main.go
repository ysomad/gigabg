package main

import (
	"context"
	"flag"
	"log/slog"
	"math/rand/v2"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ysomad/gigabg/client"
	"github.com/ysomad/gigabg/config"
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/game/catalog"
	"github.com/ysomad/gigabg/ui"
	"github.com/ysomad/gigabg/ui/scene"
	"github.com/ysomad/gigabg/ui/widget"
)

func main() {
	configPath := flag.String("config", "", "path to client config TOML file")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))

	if *configPath == "" {
		slog.Error("config flag is required")
		os.Exit(1)
	}

	cfg, err := config.LoadClient(*configPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	slog.Info("config loaded", "path", *configPath)

	cards, err := catalog.New()
	if err != nil {
		slog.Error("card catalog", "error", err)
		return
	}

	app, err := ui.NewApp()
	if err != nil {
		slog.Error("init app failed", "error", err)
		return
	}
	app.SetDebug(cfg.Dev.Debug)

	httpClient := client.New(cfg.Server.Addr, cfg.Server.Proxy)

	w := float64(ui.BaseWidth)
	h := float64(ui.BaseHeight)
	popupW := w * 0.40
	popupH := h * 0.25
	popupRect := ui.Rect{X: w/2 - popupW/2, Y: h/2 - popupH/2, W: popupW, H: popupH}

	var showMenu func()

	connectAndPlay := func(p *widget.Popup, player game.PlayerID, lobbyID string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		p.SetMessage("Connecting to server...")
		slog.Info("connecting", "player", player, "lobby", lobbyID)

		gc, err := client.NewGameClient(ctx, cfg.Server.Addr, player, lobbyID, cfg.Server.Proxy)
		if err != nil {
			slog.Error("connection failed", "error", err)
			p.SetTitle("Error")
			p.SetMessage(err.Error())
			p.ShowButton("Close", func() { app.HideOverlay() })
			return
		}

		p.SetMessage("Waiting for game state...")
		slog.Info("waiting for state", "player", player, "lobby", lobbyID)

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

		slog.Info("joined game", "player", player, "lobby", lobbyID)
		app.HideOverlay()
		app.SwitchScene(scene.NewGame(gc, cards, app.Font(), app.BoldFont(), func() {
			if err := gc.Close(); err != nil {
				slog.Error("close game client", "error", err)
			}
			showMenu()
		}))
	}

	onJoin := func(player game.PlayerID, lobbyID string) {
		p := widget.NewPopup(app.Font(), popupRect, "", "Connecting...")
		app.ShowOverlay(p)

		go func() {
			connectAndPlay(p, player, lobbyID)
		}()
	}

	onCreate := func(player game.PlayerID, lobbySize int) {
		p := widget.NewPopup(app.Font(), popupRect, "", "Creating lobby...")
		app.ShowOverlay(p)

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			slog.Info("creating lobby", "player", player, "size", lobbySize)

			lobbyID, err := httpClient.CreateLobby(ctx, lobbySize)
			if err != nil {
				slog.Error("create lobby failed", "error", err)
				p.SetTitle("Error")
				p.SetMessage(err.Error())
				p.ShowButton("Close", func() { app.HideOverlay() })
				return
			}

			slog.Info("lobby created", "lobby", lobbyID)
			connectAndPlay(p, player, lobbyID)
		}()
	}

	showMenu = func() {
		app.SwitchScene(scene.NewMenu(app.Font(), onJoin, onCreate))
	}
	showMenu()

	if cfg.Dev.Lobby != "" {
		onJoin(game.PlayerID(rand.Int32N(100_000_000)), cfg.Dev.Lobby)
	}

	ebiten.SetWindowSize(ui.BaseWidth, ui.BaseHeight)
	ebiten.SetWindowTitle("GIGA Battlegrounds")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(app); err != nil {
		slog.Error("run game failed", "error", err)
		os.Exit(1)
	}
}
