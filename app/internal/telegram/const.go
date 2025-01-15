package telegram

const (
	EmojiComputer = "🖥️"
	EmojiFree     = "✅"
	EmojiBusy     = "❌"

	ErrNoEnvironments    = "no environments found"
	ErrNoBusyStands      = "no busy stands found"
	ErrNoFreeStands      = "no free stands available"
	ErrStandBusy         = "stand is busy, choose another free one"
	ErrStandNotFound     = "stand not found"
	ErrNoStandsToRelease = "you have no stands to release"
	ErrFailedToClaim     = "failed to claim stand: %v"
	ErrFailedToRelease   = "failed to release stand: %v"

	MsgChooseStand      = "сhoose stand to claim:"
	MsgChooseToRelease  = "сhoose stand to release:"
	MsgChooseUserToPing = "сhoose user to ping:"

	TplStandClaimed  = "@%s has claimed %s"
	TplStandReleased = "@%s has released %s"
	TplPingUser      = "@%s would you mind releasing your stands??"
	TplPingAllUsers  = "%s, would you mind releasing your stands?"
	TplStandBusyBy   = "busy by @%s for %d h. %s"
	TplStandFree     = "is free %s"
	TplGreetings     = "Hello @%s, I'm StandClaimer bot, I will help you to manage environments across the team. Tap `/` on the group menu to see commands"
	TplStandInfo     = "%s %s %s"
	TplUserStand     = "@%s: %s"
	TplButtonStand   = "%s %s"
	TplButtonUser    = "@%s (%s)"
	TplFeatureState  = "feature: %s %s"
)
