package env

import (
	"reflect"
	"testing"
)

func TestString(t *testing.T) {
	const key = "UAPP_TEST_ENV_STRING"

	t.Setenv(key, "")
	if got := String(key, "def"); got != "def" {
		t.Fatalf("空值应返回默认值, got %q", got)
	}

	t.Setenv(key, "value")
	if got := String(key, "def"); got != "value" {
		t.Fatalf("应返回环境变量值, got %q", got)
	}
}

func TestBool(t *testing.T) {
	const key = "UAPP_TEST_ENV_BOOL"

	t.Setenv(key, "")
	if got := Bool(key, true); !got {
		t.Fatal("空值应返回默认值 true")
	}

	t.Setenv(key, "false")
	if got := Bool(key, true); got {
		t.Fatal("false 应解析为 false")
	}

	t.Setenv(key, "1")
	if got := Bool(key, false); !got {
		t.Fatal("1 应解析为 true")
	}

	t.Setenv(key, "not-a-bool")
	if got := Bool(key, true); !got {
		t.Fatal("非法值应返回默认值")
	}
}

func TestList(t *testing.T) {
	const key = "UAPP_TEST_ENV_LIST"

	t.Setenv(key, "")
	if got := List(key); got != nil {
		t.Fatalf("空值应返回 nil, got %#v", got)
	}

	t.Setenv(key, "a, b ,,c")
	want := []string{"a", "b", "c"}
	if got := List(key); !reflect.DeepEqual(got, want) {
		t.Fatalf("列表解析错误: got %#v, want %#v", got, want)
	}
}
