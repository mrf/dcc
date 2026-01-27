package ui

import "testing"

func TestTruncateShortString(t *testing.T) {
	result := Truncate("hello", 10)
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}

func TestTruncateExactLength(t *testing.T) {
	result := Truncate("hello", 5)
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}

func TestTruncateLongString(t *testing.T) {
	result := Truncate("hello world", 8)
	if result != "hello..." {
		t.Errorf("expected 'hello...', got '%s'", result)
	}
}

func TestTruncateVeryShortMax(t *testing.T) {
	result := Truncate("hello", 3)
	if result != "hel" {
		t.Errorf("expected 'hel', got '%s'", result)
	}
}

func TestAgeColorGreen(t *testing.T) {
	if AgeColor(0) != ColorGreen {
		t.Error("expected green for 0 days")
	}
	if AgeColor(1) != ColorGreen {
		t.Error("expected green for 1 day")
	}
}

func TestAgeColorYellow(t *testing.T) {
	if AgeColor(2) != ColorYellow {
		t.Error("expected yellow for 2 days")
	}
	if AgeColor(4) != ColorYellow {
		t.Error("expected yellow for 4 days")
	}
}

func TestAgeColorOrange(t *testing.T) {
	if AgeColor(5) != ColorOrange {
		t.Error("expected orange for 5 days")
	}
	if AgeColor(6) != ColorOrange {
		t.Error("expected orange for 6 days")
	}
}

func TestAgeColorRed(t *testing.T) {
	if AgeColor(7) != ColorRed {
		t.Error("expected red for 7 days")
	}
	if AgeColor(100) != ColorRed {
		t.Error("expected red for 100 days")
	}
}

func TestStashAgeColorGray(t *testing.T) {
	if StashAgeColor(0) != ColorGray {
		t.Error("expected gray for 0 days")
	}
	if StashAgeColor(6) != ColorGray {
		t.Error("expected gray for 6 days")
	}
}

func TestStashAgeColorYellow(t *testing.T) {
	if StashAgeColor(7) != ColorYellow {
		t.Error("expected yellow for 7 days")
	}
	if StashAgeColor(29) != ColorYellow {
		t.Error("expected yellow for 29 days")
	}
}

func TestStashAgeColorRed(t *testing.T) {
	if StashAgeColor(30) != ColorRed {
		t.Error("expected red for 30 days")
	}
	if StashAgeColor(100) != ColorRed {
		t.Error("expected red for 100 days")
	}
}

func TestPortColorWebServers(t *testing.T) {
	if PortColor(80) != ColorGreen {
		t.Error("expected green for port 80")
	}
	if PortColor(443) != ColorGreen {
		t.Error("expected green for port 443")
	}
	if PortColor(8080) != ColorGreen {
		t.Error("expected green for port 8080")
	}
	if PortColor(8443) != ColorGreen {
		t.Error("expected green for port 8443")
	}
}

func TestPortColorDatabases(t *testing.T) {
	if PortColor(5432) != ColorYellow {
		t.Error("expected yellow for port 5432 (postgres)")
	}
	if PortColor(3306) != ColorYellow {
		t.Error("expected yellow for port 3306 (mysql)")
	}
	if PortColor(27017) != ColorYellow {
		t.Error("expected yellow for port 27017 (mongo)")
	}
	if PortColor(6379) != ColorYellow {
		t.Error("expected yellow for port 6379 (redis)")
	}
}

func TestPortColorDevServers(t *testing.T) {
	if PortColor(3000) != ColorCyan {
		t.Error("expected cyan for port 3000")
	}
	if PortColor(3999) != ColorCyan {
		t.Error("expected cyan for port 3999")
	}
}

func TestPortColorOther(t *testing.T) {
	if PortColor(22) != ColorWhite {
		t.Error("expected white for port 22")
	}
	if PortColor(9999) != ColorWhite {
		t.Error("expected white for port 9999")
	}
}
