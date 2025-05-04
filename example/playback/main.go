package main

import (
	"bytes"
	replay "github.com/akmalfairuz/df-replay"
	"github.com/bedrock-gophers/intercept/intercept"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
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
		_ = p.Inventory().SetItem(1, item.NewStack(item.AmethystShard{}, 1).WithValue("replay_item", "fast_forward").WithCustomName("Fast Forward"))
		_ = p.Inventory().SetItem(2, item.NewStack(item.EchoShard{}, 1).WithValue("replay_item", "rewind").WithCustomName("Rewind"))
		_ = p.Inventory().SetItem(3, item.NewStack(item.Emerald{}, 1).WithValue("replay_item", "speed_up").WithCustomName("Speed Up"))
		_ = p.Inventory().SetItem(4, item.NewStack(item.GoldIngot{}, 1).WithValue("replay_item", "speed_down").WithCustomName("Speed Down"))
		playback.Play()
		p.SetGameMode(world.GameModeSpectator)
		srv.World().SetTime(3000)
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
		h.p.SetReverse(!h.p.Reversed())
		ctx.Val().Messagef("reverse: %v", h.p.Reversed())
	case "fast_forward":
		h.p.FastForward(ctx.Val().Tx(), 100) // 5seconds
	case "rewind":
		h.p.Rewind(ctx.Val().Tx(), 100) // 5seconds
	case "speed_up":
		// Increase playback speed by 0.5, up to a maximum of 5.0
		newSpeed := min(h.p.Speed()+0.5, 5.0)
		h.p.SetSpeed(newSpeed)
		ctx.Val().Messagef("playback speed: %.1fx", newSpeed)
	case "speed_down":
		// Decrease playback speed by 0.5, down to a minimum of 0.1
		newSpeed := max(h.p.Speed()-0.5, 0.1)
		h.p.SetSpeed(newSpeed)
		ctx.Val().Messagef("playback speed: %.1fx", newSpeed)
	}
}
