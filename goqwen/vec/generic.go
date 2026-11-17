package vec

// Переносимые реализации ядер на чистом Go. Используются на архитектурах без
// ассемблерных версий, на старых процессорах и как эталон в тестах.

func dotGeneric(a, b []float32) float32 {
	b = b[:len(a)]
	var s0, s1, s2, s3 float32
	i := 0
	for ; i+4 <= len(a); i += 4 {
		s0 += a[i] * b[i]
		s1 += a[i+1] * b[i+1]
		s2 += a[i+2] * b[i+2]
		s3 += a[i+3] * b[i+3]
	}
	for ; i < len(a); i++ {
		s0 += a[i] * b[i]
	}
	return (s0 + s1) + (s2 + s3)
}

func axpyGeneric(alpha float32, x, y []float32) {
	y = y[:len(x)]
	for i, v := range x {
		y[i] += alpha * v
	}
}

// dotQ8Generic: строка в формате Q8_0 (блоки по 34 байта: масштаб float16 и
// 32 значения int8), скалярно умноженная на x.
func dotQ8Generic(row []byte, x []float32) float32 {
	var total float32
	for len(x) >= 32 && len(row) >= 34 {
		d := f16Table[uint16(row[0])|uint16(row[1])<<8]
		q := row[2:34]
		xb := x[:32]
		var s0, s1 float32
		for j := 0; j < 32; j += 2 {
			s0 += float32(int8(q[j])) * xb[j]
			s1 += float32(int8(q[j+1])) * xb[j+1]
		}
		total += d * (s0 + s1)
		row = row[34:]
		x = x[32:]
	}
	return total
}
