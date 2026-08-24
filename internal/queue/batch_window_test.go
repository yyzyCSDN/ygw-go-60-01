package queue

import (
	"strconv"
	"testing"

	"hookrelay/internal/model"
)

// enqueueN 向队列追加 n 条事件并返回。每条事件使用不同的请求体，
// 以保证由内容散列派生的事件 ID 互不相同。
func enqueueN(t *testing.T, q *Queue, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		ev := model.NewEvent("", "order.created", []byte(`{"i":`+strconv.Itoa(i)+`}`))
		if _, err := q.Enqueue(ev); err != nil {
			t.Fatalf("enqueue %d: %v", i, err)
		}
	}
}

// TestBatchWindowBoundariesAreContinuousAndNonOverlapping 验证连续两次批量
// 窗口切分不重叠、不漏事件：第一次窗口投递后位点推进到窗口末尾，
// 第二次窗口必须从末尾的下一条开始，既不重投边界事件，也不跳号。
func TestBatchWindowBoundariesAreContinuousAndNonOverlapping(t *testing.T) {
	q := NewQueue()
	enqueueN(t, q, 5) // seq 1..5

	// 窗口 1：after=0, size=2，应返回 seq 1,2。
	w1 := q.Batch(0, 2)
	if len(w1) != 2 || w1[0].Seq != 1 || w1[1].Seq != 2 {
		t.Fatalf("window1 expected seq 1,2, got %v", seqs(w1))
	}

	// 位点推进到窗口 1 的末尾 seq=2。
	after := w1[len(w1)-1].Seq

	// 窗口 2：after=2, size=2，应返回 seq 3,4，绝不能重投 seq 2。
	w2 := q.Batch(after, 2)
	if len(w2) != 2 || w2[0].Seq != 3 || w2[1].Seq != 4 {
		t.Fatalf("window2 expected seq 3,4, got %v (window boundary re-delivered seq %d)",
			seqs(w2), after)
	}
}

// TestBatchExcludesDeliveredOffset 验证位点指向的事件不会被下一个
// 窗口重复包含。这是批量窗口边界"算重"导致重复投递的直接复现。
func TestBatchExcludesDeliveredOffset(t *testing.T) {
	q := NewQueue()
	enqueueN(t, q, 3) // seq 1,2,3

	// 假设 seq 2 已投递并确认，位点 = 2。
	got := q.Batch(2, 2)
	for _, ev := range got {
		if ev.Seq == 2 {
			t.Fatalf("delivered offset seq 2 must be excluded from next window, got %v",
				seqs(got))
		}
	}
	if len(got) != 1 || got[0].Seq != 3 {
		t.Fatalf("expected only seq 3 after offset 2, got %v", seqs(got))
	}
}

// TestResumeFromExcludesConfirmedOffset 验证续投从位点下一条开始，
// 不重投已确认的位点事件。
func TestResumeFromExcludesConfirmedOffset(t *testing.T) {
	q := NewQueue()
	enqueueN(t, q, 3) // seq 1,2,3

	got := q.ResumeFrom(2)
	for _, ev := range got {
		if ev.Seq == 2 {
			t.Fatalf("confirmed offset seq 2 must be excluded from resume, got %v",
				seqs(got))
		}
	}
	if len(got) != 1 || got[0].Seq != 3 {
		t.Fatalf("expected resume to return only seq 3 after offset 2, got %v", seqs(got))
	}
}

func seqs(events []*model.Event) []uint64 {
	out := make([]uint64, 0, len(events))
	for _, ev := range events {
		out = append(out, ev.Seq)
	}
	return out
}
