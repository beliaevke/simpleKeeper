package app

import (
	"github.com/beliaevke/simpleKeeper/client/models"

	"github.com/rivo/tview"
)

func FormMain(ac *AppClient) *tview.Form {

	form := tview.NewForm().
		AddTextView("Greetings!", "Login or register to access:", 40, 1, true, false).
		AddInputField("Login", "", 20, nil, func(text string) {
			ac.Client.UserLogPass.UserLogin = text
		}).
		AddPasswordField("Password", "", 20, '*', func(text string) {
			ac.Client.UserLogPass.UserPassword = text
		}).
		AddButton("Login", func() {
			if success, errMsg := ac.Login(); success {
				ac.Client.Pages.AddAndSwitchToPage("FormAuthMain", FormAuthMain(ac), true)
			} else {
				ac.Client.Pages.AddAndSwitchToPage("FormLoginErr", FormLoginErr(ac, errMsg), true)
			}
		}).
		AddButton("Register", func() {
			if success, errMsg := ac.Register(); success {
				ac.Client.Pages.AddAndSwitchToPage("FormAuthMain", FormAuthMain(ac), true)
			} else {
				ac.Client.Pages.AddAndSwitchToPage("FormLoginErr", FormLoginErr(ac, errMsg), true)
			}
		}).
		AddButton("Quit", func() {
			ac.Client.App.Stop()
		})

	form.SetBorder(true).SetTitle("Simple Keeper").SetTitleAlign(tview.AlignLeft)

	return form
}

func FormAuthMain(ac *AppClient) *tview.Form {

	// TODO check token

	welcomeText :=
		`		Welcome back!
Add a new secret or view the list ↓
	`
	form := tview.NewForm().
		AddTextView("ver.:", ac.Client.Cfg.FlagVersion, 40, 1, true, false).
		AddTextView("", welcomeText, 40, 3, true, false).
		AddButton("Add Secret", func() {
			ac.Client.Pages.AddAndSwitchToPage("FormAddSecret", FormAddSecret(ac), true)
		}).
		AddButton("Secrets List", func() {
			ac.Client.Pages.AddAndSwitchToPage("FormSecretsList", FormSecretsList(ac), true)
		}).
		AddButton("Quit", func() {
			ac.Client.App.Stop()
		})

	form.SetBorder(true).SetTitle("Simple Keeper").SetTitleAlign(tview.AlignLeft)

	return form
}

func FormError(ac *AppClient, errMsq string) *tview.Form {

	form := tview.NewForm().
		AddTextView("ERROR!", errMsq, 80, 5, true, false).
		AddButton("Quit", func() {
			ac.Client.App.Stop()
		})

	form.SetBorder(true).SetTitle("Simple Keeper").SetTitleAlign(tview.AlignLeft)

	return form
}

func FormAddSecret(ac *AppClient) *tview.Form {

	var secretType int

	form := tview.NewForm().
		AddDropDown("Select type", []string{"LOGPASS", "TEXT", "BINARY", "BCARD"}, 0, func(option string, optionIndex int) {
			secretType = optionIndex
		}).
		AddButton("Next", func() {
			switch secretType {
			case 0:
				ac.Client.Pages.AddAndSwitchToPage("FormAddLogPass", FormAddLogPass(ac), true)
			case 1:
				ac.Client.Pages.AddAndSwitchToPage("FormAddText", FormAddText(ac), true)
			case 2:
				//ac.Client.Pages.AddAndSwitchToPage("FormAddBinary", FormAddBinary(ac), true)
				ac.Client.Pages.AddAndSwitchToPage("FormAuthMain", FormAuthMain(ac), true)
			case 3:
				ac.Client.Pages.AddAndSwitchToPage("FormAddBCard", FormFormAddBCard(ac), true)
			default:
				ac.Client.Pages.AddAndSwitchToPage("FormAuthMain", FormAuthMain(ac), true)
			}
		}).
		AddButton("Back", func() {
			ac.Client.Pages.AddAndSwitchToPage("FormAuthMain", FormAuthMain(ac), true)
		}).
		AddButton("Quit", func() {
			ac.Client.App.Stop()
		})

	form.SetBorder(true).SetTitle("Add Secret").SetTitleAlign(tview.AlignLeft)

	return form
}

func FormSecretsList(ac *AppClient) *tview.Form {

	var SelSecret string
	secretsList, err := ac.GetSecretsList()
	if err != nil {
		ac.Client.Pages.AddAndSwitchToPage("FormError", FormError(ac, err.Error()), true)
	}

	form := tview.NewForm().
		AddTextView("LogPass:", secretsList.logpass, 80, 1, true, false).
		AddTextView("BCard:", secretsList.bcard, 80, 1, true, false).
		AddTextView("Text:", secretsList.text, 80, 1, true, false).
		AddTextView("Binary:", secretsList.binary, 80, 1, true, false).
		AddDropDown("Select secret", secretsList.names, 0, func(option string, optionIndex int) {
			SelSecret = option
		}).
		AddButton("Open secret", func() {
			if content, err := ac.GetSecret(SelSecret); err == nil {
				ac.Client.Pages.AddAndSwitchToPage("FormViewSecret", FormViewSecret(ac, SelSecret, content), true)
			} else {
				ac.Client.Pages.AddAndSwitchToPage("FormError", FormError(ac, err.Error()), true)
			}
		}).
		AddButton("Back", func() {
			ac.Client.Pages.AddAndSwitchToPage("FormAuthMain", FormAuthMain(ac), true)
		}).
		AddButton("Quit", func() {
			ac.Client.App.Stop()
		})

	form.SetBorder(true).SetTitle("Secrets List").SetTitleAlign(tview.AlignLeft)

	return form
}

func FormAddLogPass(ac *AppClient) *tview.Form {

	var name, addSecretErrMsg string
	var logpass models.LogPass

	form := tview.NewForm().
		AddInputField("Secret name", "", 20, nil, func(text string) {
			name = text
		}).
		AddInputField("Login", "", 20, nil, func(text string) {
			logpass.Login = text
		}).
		AddInputField("Password", "", 20, nil, func(text string) {
			logpass.Password = text
		}).
		AddButton("Add", func() {
			var emptyFields bool
			if name == "" {
				addSecretErrMsg = "Secret name is empty"
				emptyFields = true
			} else if logpass.Login == "" {
				addSecretErrMsg = "Login is empty"
				emptyFields = true
			} else if logpass.Password == "" {
				addSecretErrMsg = "Password is empty"
				emptyFields = true
			}
			if emptyFields {
				ac.Client.Pages.AddAndSwitchToPage("FormAddSecretErr", FormAddSecretErr(ac, "FormAddLogPass", addSecretErrMsg), true)
			} else {
				if success, errMsg := ac.AddLogPass(name, logpass); success {
					ac.Client.Pages.AddAndSwitchToPage("FormAuthMain", FormAuthMain(ac), true)
				} else {
					ac.Client.Pages.AddAndSwitchToPage("FormAddSecretErr", FormAddSecretErr(ac, "FormAddLogPass", errMsg), true)
				}
			}
		}).
		AddButton("Back", func() {
			ac.Client.Pages.AddAndSwitchToPage("FormAuthMain", FormAuthMain(ac), true)
		}).
		AddButton("Quit", func() {
			ac.Client.App.Stop()
		})

	form.SetBorder(true).SetTitle("Add Secret").SetTitleAlign(tview.AlignLeft)

	return form
}

func FormAddText(ac *AppClient) *tview.Form {

	var name, addSecretErrMsg string
	var txt models.Text

	form := tview.NewForm().
		AddInputField("Secret name", "", 20, nil, func(text string) {
			name = text
		}).
		AddInputField("TEXT", "", 20, nil, func(text string) {
			txt.Data = text
		}).
		AddButton("Add", func() {
			var emptyFields bool
			if name == "" {
				addSecretErrMsg = "Secret name is empty"
				emptyFields = true
			} else if txt.Data == "" {
				addSecretErrMsg = "Text is empty"
				emptyFields = true
			}
			if emptyFields {
				ac.Client.Pages.AddAndSwitchToPage("FormAddSecretErr", FormAddSecretErr(ac, "FormAddText", addSecretErrMsg), true)
			} else {
				if success, errMsg := ac.AddText(name, txt); success {
					ac.Client.Pages.AddAndSwitchToPage("FormAuthMain", FormAuthMain(ac), true)
				} else {
					ac.Client.Pages.AddAndSwitchToPage("FormAddSecretErr", FormAddSecretErr(ac, "FormAddText", errMsg), true)
				}
			}
		}).
		AddButton("Back", func() {
			ac.Client.Pages.AddAndSwitchToPage("FormAuthMain", FormAuthMain(ac), true)
		}).
		AddButton("Quit", func() {
			ac.Client.App.Stop()
		})

	form.SetBorder(true).SetTitle("Add Secret").SetTitleAlign(tview.AlignLeft)

	return form
}

func FormFormAddBCard(ac *AppClient) *tview.Form {

	var name, addSecretErrMsg string
	var bcard models.BCard

	form := tview.NewForm().
		AddInputField("Secret name", "", 20, nil, func(text string) {
			name = text
		}).
		AddInputField("Number", "", 20, nil, func(text string) {
			bcard.Number = text
		}).
		AddInputField("Holder", "", 20, nil, func(text string) {
			bcard.Holder = text
		}).
		AddInputField("Expiry date", "", 10, nil, func(text string) {
			bcard.ExpiryDate = text
		}).
		AddInputField("Security code", "", 10, nil, func(text string) {
			bcard.SecurityCode = text
		}).
		AddButton("Add", func() {
			var emptyFields bool
			if name == "" {
				addSecretErrMsg = "Secret name is empty"
				emptyFields = true
			} else if bcard.Number == "" {
				addSecretErrMsg = "Number is empty"
				emptyFields = true
			} else if bcard.Holder == "" {
				addSecretErrMsg = "Holder is empty"
				emptyFields = true
			} else if bcard.ExpiryDate == "" {
				addSecretErrMsg = "Expiry date is empty"
				emptyFields = true
			}
			if emptyFields {
				ac.Client.Pages.AddAndSwitchToPage("FormAddSecretErr", FormAddSecretErr(ac, "FormFormAddBCard", addSecretErrMsg), true)
			} else {
				if success, errMsg := ac.AddBCard(name, bcard); success {
					ac.Client.Pages.AddAndSwitchToPage("FormAuthMain", FormAuthMain(ac), true)
				} else {
					ac.Client.Pages.AddAndSwitchToPage("FormAddSecretErr", FormAddSecretErr(ac, "FormFormAddBCard", errMsg), true)
				}
			}
		}).
		AddButton("Back", func() {
			ac.Client.Pages.AddAndSwitchToPage("FormAuthMain", FormAuthMain(ac), true)
		}).
		AddButton("Quit", func() {
			ac.Client.App.Stop()
		})

	form.SetBorder(true).SetTitle("Add Secret").SetTitleAlign(tview.AlignLeft)

	return form
}

func FormViewSecret(ac *AppClient, name string, content string) *tview.Form {

	secret, err := models.DecodeSecret(content)
	if err != nil {
		return FormError(ac, "Client ViewLogPass models.DecodeSecret error: "+err.Error())
	}

	form := tview.NewForm().
		AddTextView("Secret name", name, 40, 1, true, false).
		AddTextView("Info", secret.String(), 100, 2, true, false).
		AddButton("Back", func() {
			ac.Client.Pages.AddAndSwitchToPage("FormSecretsList", FormSecretsList(ac), true)
		}).
		AddButton("(!) Delete secret", func() {
			if success, err := ac.DeleteSecret(name); success {
				ac.Client.Pages.AddAndSwitchToPage("FormAuthMain", FormAuthMain(ac), true)
			} else {
				ac.Client.Pages.AddAndSwitchToPage("FormError", FormError(ac, "Client deleteSecret error: "+err.Error()), true)
			}
		}).
		AddButton("Quit", func() {
			ac.Client.App.Stop()
		})

	form.SetBorder(true).SetTitle("Secret").SetTitleAlign(tview.AlignLeft)

	return form
}

func FormLoginErr(ac *AppClient, errMsg string) *tview.Form {

	form := tview.NewForm().
		AddTextView("", errMsg, 80, 2, true, false).
		AddButton("Try again", func() {
			ac.Client.Pages.SwitchToPage("FormMain")
		}).
		AddButton("Quit", func() {
			ac.Client.App.Stop()
		})

	form.SetBorder(true).SetTitle("Simple Keeper").SetTitleAlign(tview.AlignLeft)

	return form
}

func FormAddSecretErr(ac *AppClient, PrevPage string, errMsg string) *tview.Form {

	form := tview.NewForm().
		AddTextView("", errMsg, 80, 2, true, false).
		AddButton("Try again", func() {
			ac.Client.Pages.SwitchToPage(PrevPage)
		}).
		AddButton("Quit", func() {
			ac.Client.App.Stop()
		})

	form.SetBorder(true).SetTitle("Add Secret").SetTitleAlign(tview.AlignLeft)

	return form
}
