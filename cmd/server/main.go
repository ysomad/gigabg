package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/game/catalog"
	"github.com/ysomad/gigabg/lobby"
	"github.com/ysomad/gigabg/pkg/httpserver"
	"github.com/ysomad/gigabg/server"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	port := flag.Int("port", 8080, "server port")
	devLobby := flag.String("dev-lobby", "", "create a 2-player dev lobby with this ID on start")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))

	ctx, notifyCancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	defer notifyCancel()

	cards, err := catalog.New()
	if err != nil {
		return fmt.Errorf("card catalog: %w", err)
	}

	memstore := lobby.NewMemoryStore()

	if *devLobby != "" {
		if err := createDevLobby(memstore, cards, *devLobby); err != nil {
			return fmt.Errorf("dev lobby: %w", err)
		}
	}

	gameServer := server.New(memstore, cards)

	srv := httpserver.New(ctx, gameServer, httpserver.WithPort(*port))

	select {
	case err := <-srv.Notify():
		slog.ErrorContext(ctx, "httpserver: "+err.Error())
	case <-ctx.Done():
		slog.InfoContext(ctx, "root context done")
	}

	if err := srv.Shutdown(ctx); err != nil {
		slog.WarnContext(ctx, "httpserver: shutdown: "+err.Error())
	}

	return nil
}

func createDevLobby(store *lobby.MemoryStore, cards game.CardCatalog, id string) error {
	l, err := lobby.New(cards, 2)
	if err != nil {
		return err
	}
	l.SetID(id)
	if err := store.CreateLobby(l); err != nil {
		return err
	}
	slog.Info("dev lobby created", "id", id)
	return nil
}
