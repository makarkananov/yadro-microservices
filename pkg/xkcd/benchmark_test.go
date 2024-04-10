package xkcd

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func BenchmarkGetComics(b *testing.B) {
	client := NewClient("https://xkcd.com")

	for _, goroutinesLimit := range []int{1, 2, 5, 10, 20, 30, 40, 50, 100, 200} {
		for j := 0; j < 10; j++ {
			b.Run(fmt.Sprintf("goroutinesLimit=%d", goroutinesLimit), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_, err := client.GetComics(context.Background(), 1000, map[int]bool{}, goroutinesLimit, 2)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
		time.Sleep(60 * time.Minute)
	}
}
