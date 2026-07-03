package main

import (
	"context"
	"fmt"

	"github.com/ihsan-ramadhan/tuckify/internal/store"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	schedules, err := store.Load()
	if err != nil {
		return fmt.Sprintf("Error loading schedules: %v", err)
	}
	return fmt.Sprintf("Hello, you have %d schedules!", len(schedules))
}
