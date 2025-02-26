package main

import (
	"context"

	"github.com/Burmuley/ovoo/internal/controllers/milter"
)

func main() {
	ovooClient := milter.NewOvooClient("http://127.0.0.1:8808")
	ctrl, _ := milter.New("127.0.0.1:8825", nil, ovooClient)
	ctrl.Start(context.Background())
}
