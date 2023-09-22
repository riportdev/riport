package servicemanagement

import (
	"context"
	"log"
	"path/filepath"

	"github.com/kardianos/service"

	chclient "github.com/riportdev/riport/client"
	chshare "github.com/riportdev/riport/share"
)

var svcConfig = &service.Config{
	Name:        "riport",
	DisplayName: "Rport Client",
	Description: "Create reverse tunnels with ease.",
}

func HandleSvcCommand(svcCommand string, configPath string, user string) error {
	svc, err := getService(nil, configPath, user)
	if err != nil {
		return err
	}

	return chshare.HandleServiceCommand(svc, svcCommand)
}

func RunAsService(c *chclient.Client, configPath string) error {
	svc, err := getService(c, configPath, "")
	if err != nil {
		return err
	}

	return svc.Run()
}

func getService(c *chclient.Client, configPath string, user string) (service.Service, error) {
	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}
	svcConfig.Arguments = []string{"-c", absConfigPath}
	if user != "" {
		svcConfig.UserName = user
	}
	return service.New(&serviceWrapper{c}, svcConfig)
}

type serviceWrapper struct {
	*chclient.Client
}

func (w *serviceWrapper) Start(service.Service) error {
	if w.Client == nil {
		return nil
	}
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := w.Client.Run(ctx); err != nil {
			log.Println(err)
		}
	}()
	return nil
}

func (w *serviceWrapper) Stop(service.Service) error {
	return w.Client.Close()
}
