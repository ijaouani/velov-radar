package messages

import (
	"fmt"
)

// Bot Commands
const (
	MsgWelcome              = "👋 Welcome to Velo'v Radar!\n\nAvailable commands:\n• /start - Combine /stop & /add \n• /add - Add one or more stations. Example: /add 1337 1338\n• /remove - Remove one or more stations. Example: /remove 1337 1338\n• /list - View your monitored stations\n• /stop - Stop monitoring all stations"
	ErrAddMissingStation    = "❌ Please specify a station number. Example: /add 1337"
	ErrAddInvalidStation    = "❌ Invalid station number. Example: /add 1337"
	ErrRemoveMissingStation = "❌ Please specify a station number. Example: /remove 1337"
	ErrRemoveInvalidStation = "❌ Invalid station number. Example: /remove 1337"
	MsgUnsubscribedAll      = "✅ Unsubscribed from all stations. Example to add one: /add 1337"
	MsgNoMonitoredStations  = "ℹ️ You are not monitoring any stations right now. Example: /add 1337"
	MsgListHeader           = "📋 Your Monitored Stations:\n"
)

func MsgStationRemoved(stationID int) string {
	return fmt.Sprintf("✅ Station %d removed from your watch list.", stationID)
}

func ErrStationNotFound(stationID int) string {
	return fmt.Sprintf("❌ Station %d not found. Please check the number.", stationID)
}

func MsgMonitoringWithoutStatus(stationID int) string {
	return fmt.Sprintf("✅ Now monitoring station %d. (Current status unavailable)", stationID)
}

func MsgMonitoringWithStatus(name string, elec, meca int) string {
	return fmt.Sprintf("✅ Now monitoring station %s.\n\nCurrent availability:\n⚡ Electric: %d\n🦵 Mechanical: %d", name, elec, meca)
}

func MsgListStationUnavailable(stationID int) string {
	return fmt.Sprintf("• Station %d (status unavailable)", stationID)
}

func MsgListStation(name string, elec, meca int) string {
	return fmt.Sprintf("• %s\n  ⚡ %d | 🦵 %d", name, elec, meca)
}

// Monitor notifications
func NotifyElecEmpty(name string) string {
	return fmt.Sprintf("🔴 No more electric Velo'v at station %s.", name)
}

func NotifyElecAvailable(name string) string {
	return fmt.Sprintf("🟢 Electric Velo'v available at station %s.", name)
}

func NotifyMecaEmpty(name string) string {
	return fmt.Sprintf("🔴 No more mechanical Velo'v at station %s.", name)
}

func NotifyMecaAvailable(name string) string {
	return fmt.Sprintf("🟢 Mechanical Velo'v available at station %s.", name)
}
