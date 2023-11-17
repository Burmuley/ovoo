package main

import (
	"encoding/json"
	"fmt"
	api2 "github.com/Burmuley/domovoi/pkg/api"
	state2 "github.com/Burmuley/domovoi/pkg/state"
	"log"
	"log/slog"
	"os"
)

func main() {
	st, err := state2.NewState(state2.WithSQLite("test.db"))
	if err != nil {
		log.Fatal(err)
	}

	createAliases := []state2.Alias{
		{
			Email:            "alias1@protected.com",
			ProtectedAddress: state2.ProtectedAddress{Email: "protected1@origin.com"},
			Comment:          "test email #1",
			ServiceName:      "google.com",
			Active:           false,
		},
		{
			Email:            "alias2@protected.com",
			ProtectedAddress: state2.ProtectedAddress{Email: "protected1@origin.com"},
			Comment:          "test email #2",
			ServiceName:      "google.com",
			Active:           true,
		},
		{
			Email:            "alias3@protected.com",
			ProtectedAddress: state2.ProtectedAddress{Email: "protected2@origin.com"},
			Comment:          "test email #3",
			ServiceName:      "google.com",
			Active:           true,
		},
	}

	for _, a := range createAliases {
		al, err := st.CreateAlias(a)
		if err != nil {
			fmt.Println("create alias error: ", err)
		}
		fmt.Println(al)
	}

	if aliases, err := st.Aliases(); err != nil {
		log.Fatal("get:", err)
	} else {
		j, _ := json.MarshalIndent(aliases, "", "  ")
		fmt.Println(string(j))
	}

	peAddr := "protected6666@origin.com"
	err = st.CreateProtectedAddress(state2.ProtectedAddress{
		Email:  peAddr,
		Active: true,
	})
	if err != nil {
		fmt.Println("create protected email error:", err)
	}

	pe, ok := st.GetProtectedAddressByEmail(peAddr)

	if !ok {
		fmt.Println("no record found for ProtectedAddress")
	}

	nal := state2.Alias{
		ProtectedAddressID: pe.ID,
		Comment:            "no comments",
		ServiceName:        "no.service.yet.com",
		Email:              "alias6666@protected.com",
		Active:             true,
	}

	al, err := st.CreateAlias(nal)

	if err != nil {
		fmt.Println("error creating alias:", err)
	}

	fmt.Println("new alias: ", al)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	apiInstance := api2.NewApi(st)
	router := api2.NewRouter(apiInstance)

	err = router.Run("127.0.0.1:8808")
	if err != nil {
		log.Fatal(err)
	}
}
