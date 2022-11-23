package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"flag"
	"github.com/guoyk93/gg"
	"github.com/guoyk93/gg/ggos"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Job struct {
	Config string
	Pull   bool
	Push   bool
	Src    string
	Dst    []string
}

func (job *Job) DockerCommand(ctx context.Context, args ...string) *exec.Cmd {
	if job.Config != "" {
		args = append([]string{"--config", job.Config}, args...)
	}
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	return cmd
}

func (job *Job) Execute(ctx context.Context) (err error) {
	defer gg.Guard(&err)

	if job.Pull {
		gg.Log("- PULL: " + job.Src)
		gg.Must0(job.DockerCommand(ctx, "pull", job.Src).Run())
		gg.Log("+ PULL: " + job.Src)
	}
	for _, item := range job.Dst {
		gg.Log("-  TAG: " + item)
		gg.Must0(job.DockerCommand(ctx, "tag", job.Src, item).Run())
		gg.Log("+  TAG: " + item)
	}
	if job.Push {
		for _, item := range job.Dst {
			gg.Log("- PUSH: " + item)
			gg.Must0(job.DockerCommand(ctx, "push", item).Run())
			gg.Log("+ PUSH: " + item)
		}
	}
	return
}

func main() {
	var err error
	defer ggos.Exit(&err)
	defer gg.Guard(&err)

	var (
		optFile         string
		optSrc          string
		optDst          string
		optPull         bool
		optPush         bool
		optDockerConfig string
		optConcurrency  int
	)

	flag.StringVar(&optFile, "f", "IMAGES.txt", "images file")
	flag.IntVar(&optConcurrency, "c", 5, "concurrency")
	flag.StringVar(&optSrc, "src", "", "source registry")
	flag.StringVar(&optDst, "dst", "", "destination registries, comma seperated")
	flag.BoolVar(&optPull, "pull", false, "pull from source registry")
	flag.BoolVar(&optPush, "push", false, "push to destination registries")
	flag.StringVar(&optDockerConfig, "docker-config", "", "override docker config directory")
	flag.Parse()

	// setup dockerconfig from env
	if optDockerConfig == "" {
		envDockerConfig := strings.TrimSpace(os.Getenv("DOCKERCONFIG_BASE64"))
		if envDockerConfig != "" {
			gg.Log("generating docker config.json from environment variable $DOCKERCONFIG_BASE64")
			optDockerConfig = filepath.Join(os.TempDir(), "replika-"+strconv.FormatInt(time.Now().UnixMilli(), 10))
			gg.Must0(os.MkdirAll(optDockerConfig, 0750))
			defer os.RemoveAll(optDockerConfig)
			gg.Must0(os.WriteFile(
				filepath.Join(optDockerConfig, "config.json"),
				gg.Must(base64.StdEncoding.DecodeString(envDockerConfig)),
				0640,
			))
		}
	}

	var destinations []string
	for _, item := range strings.Split(optDst, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		destinations = append(destinations, item)
	}

	content := gg.Must(os.ReadFile(optFile))

	var jobs []Job

	for _, item := range bytes.Split(content, []byte{'\n'}) {
		name := string(bytes.TrimSpace(item))
		if name == "" {
			continue
		}
		jobs = append(jobs, Job{
			Pull: optPull,
			Push: optPush,
			Src:  path.Join(optSrc, name),
			Dst: gg.Map(destinations, func(s string) string {
				return path.Join(s, name)
			}),
		})
	}

	if optConcurrency < 1 {
		optConcurrency = 5
	}

	pg := make(chan struct{}, optConcurrency)
	for i := 0; i < optConcurrency; i++ {
		pg <- struct{}{}
	}
	wg := &sync.WaitGroup{}
	eg := gg.NewErrorGroup()

	for _i, _job := range jobs {
		var (
			i   = _i
			job = _job
		)
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-pg
			go func() {
				pg <- struct{}{}
			}()
			eg.Set(i, job.Execute(context.Background()))
		}()
	}

	wg.Wait()

	err = eg.Unwrap()
}
