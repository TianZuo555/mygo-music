package app

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"

	"mygo-music/internal/music"
)

// A lazy section asks for what it holds as the page scrolls it near, shows it
// then, and asks no more: the page's first answer carries none of it.
func TestLazySectionLoadsAsScrolled(t *testing.T) {
	a := newTestApp()
	ch := posts(a)
	src := fake(a)
	tracks := make([]music.Track, 30)
	for i := range tracks {
		tracks[i] = song(fmt.Sprint(i), fmt.Sprintf("Song %d", i))
	}
	src.browse = func(string) (*music.Browse, error) {
		return &music.Browse{Sections: []music.Section{
			{Title: "Top", Tracks: tracks},
			{Title: "Explore", ID: "explore", Lazy: true},
		}}, nil
	}
	asked := make(chan string, 2)
	src.section = func(page, section string) (*music.Section, error) {
		asked <- section
		return &music.Section{Shows: []music.Show{{ID: "s1", Name: "A Show"}}}, nil
	}
	a.open(route{page: pageBrowse, src: "fake", id: "podcasts"})
	until(t, ch, func() bool { return a.library.browses["podcasts"].page != nil })

	tt := ui.NewTester(a.view, 1200, 500)
	settled(tt)
	select {
	case s := <-asked:
		t.Fatalf("asked for %q before scrolling to it", s)
	default:
	}

	tt.Scroll(600, 250, 0, 4000)
	settled(tt)
	if s := wait(t, asked, "the section"); s != "explore" {
		t.Fatalf("asked for %q", s)
	}
	until(t, ch, func() bool { return !a.library.browses["podcasts"].page.Sections[1].Lazy })
	settled(tt)
	if _, ok := tt.Find("A Show"); !ok {
		t.Fatalf("shows %q", tt.Texts())
	}

	tt.Scroll(600, 250, 0, -4000)
	settled(tt)
	tt.Scroll(600, 250, 0, 4000)
	settled(tt)
	select {
	case s := <-asked:
		t.Fatalf("asked again for %q", s)
	default:
	}
}

// Home's lazy sections come as it scrolls to them too, named by their title
// while the page has none.
func TestLazySectionLoadsOnHome(t *testing.T) {
	a := newTestApp()
	ch := posts(a)
	src := fake(a)
	tracks := make([]music.Track, 30)
	for i := range tracks {
		tracks[i] = song(fmt.Sprint(i), fmt.Sprintf("Song %d", i))
	}
	src.home = func() (*music.Browse, error) {
		return &music.Browse{Sections: []music.Section{
			{Title: "Top", Tracks: tracks},
			{Title: "Podcasts", Lazy: true},
		}}, nil
	}
	asked := make(chan string, 2)
	src.section = func(page, section string) (*music.Section, error) {
		asked <- page + "/" + section
		return &music.Section{Shows: []music.Show{{ID: "s1", Name: "A Show"}}}, nil
	}
	a.loadHome(a.library)
	until(t, ch, func() bool { return a.library.home != nil })

	tt := ui.NewTester(a.view, 1200, 500)
	settled(tt)
	tt.Scroll(600, 250, 0, 4000)
	settled(tt)
	if got := wait(t, asked, "Home's section"); got != "/Podcasts" {
		t.Fatalf("asked for %q", got)
	}
	until(t, ch, func() bool { return !a.library.home.Sections[1].Lazy })
	settled(tt)
	if _, ok := tt.Find("A Show"); !ok {
		t.Fatalf("shows %q", tt.Texts())
	}
}
