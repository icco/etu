package main

import (
	"charm.land/huh/v2"
)

// etuTheme is huh's Charm theme with the list's help and description colors, so
// the forms and the entry list read the same. huh dims both to near-black, and
// passes through the terminal's background so the palette flips with it.
var etuTheme = huh.ThemeFunc(func(isDark bool) *huh.Styles {
	st := newStyles(isDark)
	s := huh.ThemeCharm(isDark)
	s.Help.ShortKey = st.key
	s.Help.FullKey = st.key
	s.Help.ShortDesc = st.desc
	s.Help.FullDesc = st.desc
	s.Help.ShortSeparator = st.sep
	s.Help.FullSeparator = st.sep
	s.Focused.Description = s.Focused.Description.Foreground(st.desc.GetForeground())
	s.Blurred.Description = s.Blurred.Description.Foreground(st.desc.GetForeground())
	return s
})
