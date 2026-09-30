package main

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestTasksByDate(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	s.AddTask(ctx, "u1", "2026-09-29", Task{Time: "15:00", Title: "поздняя", Kind: "other"})
	s.AddTask(ctx, "u1", "2026-09-29", Task{Time: "08:00", Title: "ранняя", Kind: "other"})
	s.AddTask(ctx, "u1", "2026-09-30", Task{Time: "09:00", Title: "завтра", Kind: "other"})
	s.AddTask(ctx, "u2", "2026-09-29", Task{Time: "10:00", Title: "чужая", Kind: "other"})

	got, err := s.TasksByDate(ctx, "u1", "2026-09-29")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Title != "ранняя" || got[1].Title != "поздняя" {
		t.Errorf("неверный список: %+v", got)
	}

	empty, _ := s.TasksByDate(ctx, "nobody", "2026-09-29")
	if empty == nil {
		t.Error("пустой список должен быть [], а не nil")
	}
}

func TestSetDone(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	task, _ := s.AddTask(ctx, "u1", "2026-09-29", Task{Time: "08:00", Title: "т", Kind: "other"})

	updated, err := s.SetDone(ctx, "u1", task.ID, true)
	if err != nil || !updated.Done {
		t.Fatalf("задача не отметилась: %+v %v", updated, err)
	}
	updated, _ = s.SetDone(ctx, "u1", task.ID, false)
	if updated.Done {
		t.Error("отметка должна сниматься")
	}

	if _, err := s.SetDone(ctx, "u1", 999, true); !errors.Is(err, ErrNotFound) {
		t.Errorf("ожидали ErrNotFound, получили %v", err)
	}
	if _, err := s.SetDone(ctx, "u2", task.ID, true); !errors.Is(err, ErrNotFound) {
		t.Errorf("чужую задачу менять нельзя: %v", err)
	}
}

func TestSetDoneConcurrent(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	task, _ := s.AddTask(ctx, "u1", "2026-09-29", Task{Time: "08:00", Title: "т", Kind: "other"})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := s.SetDone(ctx, "u1", task.ID, i%2 == 0); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()

	n, _ := s.CountTasks(ctx, "u1", "2026-09-29")
	if n != 1 {
		t.Errorf("задача должна остаться одна, а их %d", n)
	}
}

func TestSeedDemoTasksOnce(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	seedDemoTasks(ctx, s, "demo", "2026-09-29")
	seedDemoTasks(ctx, s, "demo", "2026-09-29")

	n, _ := s.CountTasks(ctx, "demo", "2026-09-29")
	if n != 5 {
		t.Errorf("демо-задачи должны добавиться один раз, а их %d", n)
	}
}

func TestProfileRoundTrip(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	empty, err := s.GetProfile(ctx, "u1")
	if err != nil || empty.Name != "" || empty.HasLocation || empty.BirthDate != "" {
		t.Fatalf("новый профиль должен быть пустым: %+v %v", empty, err)
	}

	p := Profile{Name: "Анна", BirthDate: "1956-03-12", Address: "Тверская, 7", Lat: 55.75, Lon: 37.61, HasLocation: true,
		ContactName: "Дочь", ContactPhone: "+7 999 123-45-67", Health: "аллергия"}
	if err := s.SaveProfile(ctx, "u1", p); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetProfile(ctx, "u1")
	if got != p {
		t.Errorf("сохранилось не то:\n%+v\n%+v", got, p)
	}

	p.HasLocation = false
	p.Address = ""
	p.BirthDate = ""
	s.SaveProfile(ctx, "u1", p)
	got, _ = s.GetProfile(ctx, "u1")
	if got.HasLocation || got.BirthDate != "" {
		t.Errorf("адрес и дата должны стереться: %+v", got)
	}
}

func TestSaveAddressKeepsOtherFields(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	s.SaveProfile(ctx, "u1", Profile{Name: "Анна", BirthDate: "1956-03-12"})

	lat, lon := 55.76, 37.63
	if err := s.SaveAddress(ctx, "u1", "Мясницкая, 20", &lat, &lon); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetProfile(ctx, "u1")
	if got.Name != "Анна" || got.BirthDate != "1956-03-12" || got.Address != "Мясницкая, 20" || !got.HasLocation {
		t.Errorf("адрес должен сохраниться, не задев остальное: %+v", got)
	}

	s.SaveAddress(ctx, "new-user", "Кремль", &lat, &lon)
	if got, _ := s.GetProfile(ctx, "new-user"); !got.HasLocation {
		t.Error("адрес нового пользователя тоже должен сохраняться")
	}

	s.SaveAddress(ctx, "u1", "", nil, nil)
	if got, _ := s.GetProfile(ctx, "u1"); got.HasLocation || got.Address != "" {
		t.Errorf("адрес должен стереться: %+v", got)
	}
}

func TestAICache(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	if _, ok, _ := s.CacheGet(ctx, "k"); ok {
		t.Fatal("кеш должен быть пустым")
	}
	s.CachePut(ctx, "k", "первый")
	s.CachePut(ctx, "k", "второй")
	if answer, ok, _ := s.CacheGet(ctx, "k"); !ok || answer != "второй" {
		t.Errorf("получили %q %v", answer, ok)
	}
}
