package main

import (
	"bytes"
	replay "github.com/akmalfairuz/df-replay"
	"github.com/bedrock-gophers/intercept/intercept"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/google/uuid"
	"log/slog"
	"os"
)

func main() {
	buf, err := os.ReadFile("replay_buffer.bin")
	if err != nil {
		panic(err)
	}

	conf, _ := server.DefaultConfig().Config(slog.Default())
	conf.ReadOnlyWorld = true
	conf.PlayerProvider = nil
	srv := conf.New()
	replay.Init()
	data := replay.NewData(uuid.New())
	if err := data.LoadActions(bytes.NewReader(buf)); err != nil {
		panic(err)
	}
	playback := replay.NewPlayback(srv.World(), data)

	srv.Listen()
	srv.CloseOnProgramEnd()
	for p := range srv.Accept() {
		intercept.Intercept(p)
		p.Handle(&playerHandler{p: playback})
		_ = p.Inventory().SetItem(0, item.NewStack(item.Diamond{}, 1).WithValue("replay_item", "toggle_reverse").WithCustomName("Toggle Reverse"))
		playback.Play()
	}
}

type playerHandler struct {
	player.NopHandler
	p *replay.Playback
}

func (h *playerHandler) HandleItemUse(ctx *player.Context) {
	mainHand, _ := ctx.Val().HeldItems()
	v, ok := mainHand.Value("replay_item")
	if !ok {
		return
	}
	switch v.(string) {
	case "toggle_reverse":
		h.p.SetReverse(!h.p.IsReverse())
		ctx.Val().Messagef("reverse: %v", h.p.IsReverse())
	}
}
