package main

import (
	"io"
	"io/ioutil"
	"net"
	"strings"
	"testing"
)

var quotes = `A lot of times people look at the negative side of what they feel they can't do. I always look on the positive side of what I can do. (Chuck Norris)
Nothing will work unless you do.   (Maya Angelou )
When you judge another, you do not define them, you define yourself.   (Wayne Dyer)
Work for something because it is good, not just because it stands a chance to succeed.  (Vaclav Havel)
All men have a sweetness in their life. That is what helps them go on. It is towards that they turn when they feel too worn out.  (Albert Camus)
From small beginnings come great things.
Do you want to know who you are? Don't ask. Act! Action will delineate and define you. (Thomas Jefferson)
Were here for a reason. I believe a bit of the reason is to throw little torches out to lead people through the dark.   (Whoopi Goldberg )
You can't create in a vacuum. Life gives you the material and dreams can propel new beginnings. (Byron Pulsifer)
Life is just a chance to grow a soul.  (A. Powell Davies)
Always be mindful of the kindness and not the faults of others. (Buddha)
Good thoughts are no better than good dreams, unless they be executed.  (Ralph Emerson)
You cannot have what you do not want. (John Acosta)
Everything has beauty, but not everyone sees it.   (Confucius )
Let us always meet each other with smile, for the smile is the beginning of love.   (Mother Teresa )
He who is fixed to a star does not change his mind.   (Leonardo da Vinci)
Compassion and happiness are not a sign of weakness but a sign of strength. (Dalai Lama)
It isn't what happens to us that causes us to suffer; it's what we say to ourselves about what happens. (Pema Chodron)
Being right is highly overrated. Even a stopped clock is right twice a day.
The best and most beautiful things in the world cannot be seen, nor touched... but are felt in the heart. (Helen Keller)`

var X string

func Benchmark_quoteServer_get(b *testing.B) {
	r := strings.NewReader(quotes)
	s, err := newServer(r)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		X = s.get()
	}
}

func Benchmark_quoteServer_handle(b *testing.B) {
	r := strings.NewReader(quotes)
	s, err := newServer(r)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs, cc := net.Pipe()
		s.Add(1)
		go s.handle(cs)
		_, err := io.Copy(ioutil.Discard, cc)
		cc.Close()
		if err != nil {
			b.Error(err)
		}
	}
}
func Test_quoteServer_handle(t *testing.T) {
	const want = "cofeve"
	r := strings.NewReader(want)
	s, err := newServer(r)
	if err != nil {
		t.Fatal(err)
	}
	cs, cc := net.Pipe()
	defer cc.Close()
	s.Add(1)
	go s.handle(cs)
	b, err := ioutil.ReadAll(cc)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(string(b))
	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func Test_quoteServer_load(t *testing.T) {
	s, err := newServer(strings.NewReader("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if quote := s.get(); !strings.Contains(quote, "foo") {
		t.Fatalf(`expected s.get() == "foo", got %s`, quote)
	}
	s.load(strings.NewReader("bar"))
	if quote := s.get(); !strings.Contains(quote, "bar") {
		t.Fatalf(`expected s.get() == "bar", got %s`, quote)
	}
}
