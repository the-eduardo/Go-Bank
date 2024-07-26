package mail

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/the-eduardo/Go-Bank/util"
	"testing"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	config, err := util.LoadConfig("..")
	require.NoError(t, err)
	fmt.Println(config)

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	fmt.Println(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "Test Email"
	content := fmt.Sprintf(`
	<h1>Test Email!</h1>
	<p>Hi, test the link below:</p>
	<p><a href="%v?id=%v&secret_code=%s">Verify Email</a></p>
	<p>This link will expire in 15 minutes.</p>
		`, config.HTTPServerAddress, util.RandomInt(1, 100), util.RandomString(config.SecretCodeLength))
	to := []string{"eduardo.xbox@live.com"}
	attachFiles := []string{"../app.env"}
	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)
}
