package main

import (
	"github.com/vkore/vkore/internal/api"
	"github.com/vkore/vkore/internal/store"
	"github.com/vkore/vkore/internal/vkore"
)

func main() {
	store.Init()
	vkore.Init()
	vkore.MigrateSchema()
	api.ListenAndServe()
}
