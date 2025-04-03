package main

import (
	"bytes"
	"fmt"
	replay "github.com/akmalfairuz/df-replay"
	"github.com/bedrock-gophers/intercept/intercept"
	"github.com/df-mc/dragonfly/server"
	"github.com/google/uuid"
	"log/slog"
	"os"
	"time"
)

func main() {

	conf, _ := server.DefaultConfig().Config(slog.Default())
	srv := conf.New()

	replay.Init()
	_ = os.Remove("replay_buffer.bin")

	replayID := uuid.New()
	recorder := replay.NewRecorder(replayID)
	recordPlayerHandler := replay.NewRecordPlayerHandler(recorder)
	recordWorldHandler := replay.NewRecordWorldHandler(recorder)
	srv.World().Handle(recordWorldHandler)

	go func() {
		time.AfterFunc(time.Second*45, func() {
			buf := bytes.NewBuffer(nil)
			if err := recorder.CloseAndSaveActions(buf); err != nil {
				panic(err)
			}
			if err := os.WriteFile("replay_buffer.bin", buf.Bytes(), 0644); err != nil {
				panic(err)
			}
			fmt.Println("Replay saved to buffer")
			_ = srv.Close()
		})
	}()
	recorder.StartTicking(srv.World())
	srv.Listen()
	srv.CloseOnProgramEnd()

	for p := range srv.Accept() {
		intercept.Intercept(p)
		p.Handle(recordPlayerHandler)
		recorder.AddPlayer(p)
	}
}
