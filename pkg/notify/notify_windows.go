//go:build windows

package notify

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/thesubh213/winitrix/pkg/runner"
)

var ErrUnsupported = errors.New("notifications not supported")

// Send shows a Windows toast notification.
func Send(title, message string) error {
	titleEsc := escapeXml(title)
	messageEsc := escapeXml(message)

	script := fmt.Sprintf(
		"$title='%s';$msg='%s';"+
			"[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null;"+
			"[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] > $null;"+
			"$xml = New-Object Windows.Data.Xml.Dom.XmlDocument;"+
			"$xml.LoadXml(\"<toast><visual><binding template='ToastGeneric'><text>\" + $title + \"</text><text>\" + $msg + \"</text></binding></visual></toast>\");"+
			"$toast = [Windows.UI.Notifications.ToastNotification]::new($xml);"+
			"$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('Winitrix');"+
			"$notifier.Show($toast)",
		psSingleQuoted(titleEsc),
		psSingleQuoted(messageEsc),
	)

	res := runner.RunSilent(context.Background(), "powershell", "-NoProfile", "-Command", script)
	return res.Err
}

func psSingleQuoted(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func escapeXml(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
	)
	return replacer.Replace(value)
}
