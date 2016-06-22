package engine

import (
	"bufio"
	"io"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	dockertypes "github.com/docker/engine-api/types"

	"gitlab.ricebook.net/platform/agent/common"
	"gitlab.ricebook.net/platform/agent/engine/logs"
	"gitlab.ricebook.net/platform/agent/types"
)

func (e *Engine) attach(container *types.Container) {
	transfer := e.forwards.Get(container.ID, 0)
	writer, err := logs.NewWriter(transfer, e.config.Log.Stdout)
	if err != nil {
		log.Errorf("Create log forward failed %s", err)
		return
	}

	outr, outw := io.Pipe()
	errr, errw := io.Pipe()
	go func() {
		ctx := context.Background()
		options := dockertypes.ContainerAttachOptions{
			Stream: true,
			Stdin:  false,
			Stdout: true,
			Stderr: true,
		}
		resp, err := e.docker.ContainerAttach(ctx, container.ID, options)
		if err != nil && err != httputil.ErrPersistEOF {
			log.Errorf("Log attach %s failed %s", container.ID[:7], err)
			return
		}
		defer resp.Close()
		defer outw.Close()
		defer errw.Close()
		_, err = stdcopy.StdCopy(outw, errw, resp.Reader)
		log.Debugf("Log attach %s finished", container.ID[:7])
		if err != nil {
			log.Errorf("Log attach get stream failed %s", err)
		}
	}()
	log.Debugf("Log attach %s success", container.ID[:7])
	pump := func(typ string, source io.Reader) {
		buf := bufio.NewReader(source)
		for {
			data, err := buf.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					log.Errorf("Log pump: %s %s %s", container.ID[:7], typ, err)
				}
				return
			}
			writer.Write(&types.Log{
				ID:         container.ID,
				Name:       container.Name,
				Type:       typ,
				EntryPoint: container.EntryPoint,
				Ident:      container.Ident,
				Data:       strings.TrimSuffix(data, "\n"),
				Datetime:   time.Now().Format(common.DATETIME_FORMAT),
			}, e.config.Log.Stdout)
		}
	}
	go pump("stdout", outr)
	go pump("stderr", errr)
}