package app

import (
	"testing"

	"github.com/beliaevke/simpleKeeper/client/config"

	"github.com/rivo/tview"
)

func TestFormMain(t *testing.T) {

	client := NewClient()

	ac := &AppClient{Client: client}

	form := FormMain(ac)

	if form == nil {
		t.Fatal("Expected a valid form, got nil")
	}

}

func TestFormAuthMain(t *testing.T) {
	app := tview.NewApplication()

	client := &Client{
		Pages: tview.NewPages(),
		App:   app,
		Cfg: config.ClientFlags{
			FlagVersion: "000",
		},
	}

	ac := &AppClient{Client: client}

	form := FormAuthMain(ac)

	if form == nil {
		t.Fatal("Expected a valid form, got nil")
	}
}

func TestFormError(t *testing.T) {
	app := tview.NewApplication()
	client := &Client{App: app}
	ac := &AppClient{Client: client}

	errForm := FormError(ac, "An error occurred")
	if errForm == nil {
		t.Fatal("Expected a valid error form, got nil")
	}
}

func TestFormAddSecret(t *testing.T) {
	app := tview.NewApplication()
	client := &Client{Pages: tview.NewPages(), App: app}
	ac := &AppClient{Client: client}

	form := FormAddSecret(ac)

	if form == nil {
		t.Fatal("Expected a valid form, got nil")
	}
}

/*
func TestFormSecretsList(t *testing.T) {
	app := tview.NewApplication()
	client := &Client{Pages: tview.NewPages(), App: app, NotifyCtx: context.Background()}
	ac := &AppClient{Client: client}

	client.UserLogPass.UserLogin = "testuser"
	client.UserLogPass.UserPassword = "securepassword"

	form := FormSecretsList(ac)

	if form == nil {
		t.Fatal("Expected a valid form, got nil")
	}

}
*/

func TestFormAddLogPass(t *testing.T) {
	app := tview.NewApplication()
	client := &Client{
		Pages: tview.NewPages(),
		App:   app,
	}
	ac := &AppClient{Client: client}

	form := FormAddLogPass(ac)

	if form == nil {
		t.Fatal("Expected a valid form, got nil")
	}
}

func TestFormAddText(t *testing.T) {
	app := tview.NewApplication()
	client := &Client{
		Pages: tview.NewPages(),
		App:   app,
	}
	ac := &AppClient{Client: client}

	form := FormAddText(ac)

	// Проверяем, что форма была создана
	if form == nil {
		t.Fatal("Expected a valid form, got nil")
	}
}

func TestFormAddBCard(t *testing.T) {
	app := tview.NewApplication()
	client := &Client{
		Pages: tview.NewPages(),
		App:   app,
	}
	ac := &AppClient{Client: client}

	form := FormFormAddBCard(ac)

	// Проверяем, что форма была создана
	if form == nil {
		t.Fatal("Expected a valid form, got nil")
	}
}

func TestFormViewSecret(t *testing.T) {
	app := tview.NewApplication()
	client := &Client{Pages: tview.NewPages(), App: app}
	ac := &AppClient{Client: client}

	form := FormViewSecret(ac, "MySecret", "content")
	if form == nil {
		t.Fatal("Expected a valid form, got nil")
	}
}

func TestFormLoginErr(t *testing.T) {
	app := tview.NewApplication()
	client := &Client{Pages: tview.NewPages(), App: app}
	ac := &AppClient{Client: client}

	errMsg := "Login failed"
	form := FormLoginErr(ac, errMsg)
	if form == nil {
		t.Fatal("Expected a valid form, got nil")
	}
}

func TestFormAddSecretErr(t *testing.T) {
	app := tview.NewApplication()
	client := &Client{Pages: tview.NewPages(), App: app}
	ac := &AppClient{Client: client}

	errMsg := "Error occurred while adding secret"
	PrevPage := "FormAddLogPass"
	form := FormAddSecretErr(ac, PrevPage, errMsg)
	if form == nil {
		t.Fatal("Expected a valid form, got nil")
	}
}
