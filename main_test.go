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

func BenchmarkGet(b *testing.B) {
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
func BenchmarkServe(b *testing.B) {
	r := strings.NewReader(quotes)
	s, err := newServer(r)
	if err != nil {
		b.Fatal(err)
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer l.Close()
	go s.serve(l)
	addr := l.Addr().String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			b.Error(err)
			continue
		}
		_, err = io.Copy(ioutil.Discard, conn)
		if err != nil {
			conn.Close()
			b.Error(err)
			continue
		}
		err = conn.Close()
		if err != nil {
			b.Error(err)
		}
	}
}
