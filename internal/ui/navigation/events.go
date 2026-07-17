package navigation

import ui "crds/internal/ui"

type PushEvent struct {
	From, To ui.ScreenIndex
}

type PopEvent struct {
	From, To ui.ScreenIndex
}

type ReplaceEvent struct {
	From, To ui.ScreenIndex
}

type ForwardEvent struct {
	From, To ui.ScreenIndex
}

type OverlayShownEvent struct {
	Overlay, Under ui.ScreenIndex
}

type OverlayHiddenEvent struct {
	Overlay, Under ui.ScreenIndex
}

type ModalPushEvent struct {
	From, To ui.ScreenIndex
}

type ModalPopEvent struct {
	From, To ui.ScreenIndex
}

type ResetEvent struct {
	To ui.ScreenIndex
}
