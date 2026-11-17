// Code generated from the Hugging Face file listings of the official Qwen GGUF repositories. DO NOT EDIT.

package goqwen

// Идентификаторы моделей каталога.
const (
	Qwen25_05B      = "qwen2.5-0.5b-instruct"
	Qwen25_15B      = "qwen2.5-1.5b-instruct"
	Qwen25_3B       = "qwen2.5-3b-instruct"
	Qwen25_7B       = "qwen2.5-7b-instruct"
	Qwen25_14B      = "qwen2.5-14b-instruct"
	Qwen25_32B      = "qwen2.5-32b-instruct"
	Qwen25_72B      = "qwen2.5-72b-instruct"
	Qwen25Coder_05B = "qwen2.5-coder-0.5b-instruct"
	Qwen25Coder_15B = "qwen2.5-coder-1.5b-instruct"
	Qwen25Coder_3B  = "qwen2.5-coder-3b-instruct"
	Qwen25Coder_7B  = "qwen2.5-coder-7b-instruct"
	Qwen25Coder_14B = "qwen2.5-coder-14b-instruct"
	Qwen25Coder_32B = "qwen2.5-coder-32b-instruct"
	Qwen3_06B       = "qwen3-0.6b"
	Qwen3_17B       = "qwen3-1.7b"
	Qwen3_4B        = "qwen3-4b"
	Qwen3_8B        = "qwen3-8b"
	Qwen3_14B       = "qwen3-14b"
	Qwen3_32B       = "qwen3-32b"
)

var catalogData = []ModelInfo{
	{
		ID: Qwen25_05B, Family: "Qwen2.5", Parameters: "0.5B", Repo: "Qwen/Qwen2.5-0.5B-Instruct-GGUF",
		Thinking: false, Description: "Qwen2.5 0.5B — универсальная чат-модель",
		Variants: []Variant{
			{Quant: Q2K, Size: 415182688, Files: []string{"qwen2.5-0.5b-instruct-q2_k.gguf"}},
			{Quant: Q3KM, Size: 432041824, Files: []string{"qwen2.5-0.5b-instruct-q3_k_m.gguf"}},
			{Quant: Q4_0, Size: 428730208, Files: []string{"qwen2.5-0.5b-instruct-q4_0.gguf"}},
			{Quant: Q4KM, Size: 491400032, Files: []string{"qwen2.5-0.5b-instruct-q4_k_m.gguf"}},
			{Quant: Q5_0, Size: 490475360, Files: []string{"qwen2.5-0.5b-instruct-q5_0.gguf"}},
			{Quant: Q5KM, Size: 522186592, Files: []string{"qwen2.5-0.5b-instruct-q5_k_m.gguf"}},
			{Quant: Q6K, Size: 650379104, Files: []string{"qwen2.5-0.5b-instruct-q6_k.gguf"}},
			{Quant: Q8_0, Size: 675710816, Files: []string{"qwen2.5-0.5b-instruct-q8_0.gguf"}},
			{Quant: FP16, Size: 1266425696, Files: []string{"qwen2.5-0.5b-instruct-fp16.gguf"}},
		},
	},
	{
		ID: Qwen25_15B, Family: "Qwen2.5", Parameters: "1.5B", Repo: "Qwen/Qwen2.5-1.5B-Instruct-GGUF",
		Thinking: false, Description: "Qwen2.5 1.5B — универсальная чат-модель",
		Variants: []Variant{
			{Quant: Q2K, Size: 752880160, Files: []string{"qwen2.5-1.5b-instruct-q2_k.gguf"}},
			{Quant: Q3KM, Size: 924455968, Files: []string{"qwen2.5-1.5b-instruct-q3_k_m.gguf"}},
			{Quant: Q4_0, Size: 1066227232, Files: []string{"qwen2.5-1.5b-instruct-q4_0.gguf"}},
			{Quant: Q4KM, Size: 1117320736, Files: []string{"qwen2.5-1.5b-instruct-q4_k_m.gguf"}},
			{Quant: Q5_0, Size: 1259173408, Files: []string{"qwen2.5-1.5b-instruct-q5_0.gguf"}},
			{Quant: Q5KM, Size: 1285494304, Files: []string{"qwen2.5-1.5b-instruct-q5_k_m.gguf"}},
			{Quant: Q6K, Size: 1464178720, Files: []string{"qwen2.5-1.5b-instruct-q6_k.gguf"}},
			{Quant: Q8_0, Size: 1894532128, Files: []string{"qwen2.5-1.5b-instruct-q8_0.gguf"}},
			{Quant: FP16, Size: 3560416288, Files: []string{"qwen2.5-1.5b-instruct-fp16.gguf"}},
		},
	},
	{
		ID: Qwen25_3B, Family: "Qwen2.5", Parameters: "3B", Repo: "Qwen/Qwen2.5-3B-Instruct-GGUF",
		Thinking: false, Description: "Qwen2.5 3B — универсальная чат-модель",
		Variants: []Variant{
			{Quant: Q2K, Size: 1376856480, Files: []string{"qwen2.5-3b-instruct-q2_k.gguf"}},
			{Quant: Q3KM, Size: 1724178848, Files: []string{"qwen2.5-3b-instruct-q3_k_m.gguf"}},
			{Quant: Q4_0, Size: 1997879712, Files: []string{"qwen2.5-3b-instruct-q4_0.gguf"}},
			{Quant: Q4KM, Size: 2104932768, Files: []string{"qwen2.5-3b-instruct-q4_k_m.gguf"}},
			{Quant: Q5_0, Size: 2383591840, Files: []string{"qwen2.5-3b-instruct-q5_0.gguf"}},
			{Quant: Q5KM, Size: 2438740384, Files: []string{"qwen2.5-3b-instruct-q5_k_m.gguf"}},
			{Quant: Q6K, Size: 2793410976, Files: []string{"qwen2.5-3b-instruct-q6_k.gguf"}},
			{Quant: Q8_0, Size: 3616088480, Files: []string{"qwen2.5-3b-instruct-q8_0.gguf"}},
			{Quant: FP16, Size: 6800646784, Files: []string{"qwen2.5-3b-instruct-fp16-00001-of-00002.gguf", "qwen2.5-3b-instruct-fp16-00002-of-00002.gguf"}},
		},
	},
	{
		ID: Qwen25_7B, Family: "Qwen2.5", Parameters: "7B", Repo: "Qwen/Qwen2.5-7B-Instruct-GGUF",
		Thinking: false, Description: "Qwen2.5 7B — универсальная чат-модель",
		Variants: []Variant{
			{Quant: Q2K, Size: 3015940000, Files: []string{"qwen2.5-7b-instruct-q2_k.gguf"}},
			{Quant: Q3KM, Size: 3808391072, Files: []string{"qwen2.5-7b-instruct-q3_k_m.gguf"}},
			{Quant: Q4_0, Size: 4431390848, Files: []string{"qwen2.5-7b-instruct-q4_0-00001-of-00002.gguf", "qwen2.5-7b-instruct-q4_0-00002-of-00002.gguf"}},
			{Quant: Q4KM, Size: 4683073632, Files: []string{"qwen2.5-7b-instruct-q4_k_m-00001-of-00002.gguf", "qwen2.5-7b-instruct-q4_k_m-00002-of-00002.gguf"}},
			{Quant: Q5_0, Size: 5315176576, Files: []string{"qwen2.5-7b-instruct-q5_0-00001-of-00002.gguf", "qwen2.5-7b-instruct-q5_0-00002-of-00002.gguf"}},
			{Quant: Q5KM, Size: 5444831360, Files: []string{"qwen2.5-7b-instruct-q5_k_m-00001-of-00002.gguf", "qwen2.5-7b-instruct-q5_k_m-00002-of-00002.gguf"}},
			{Quant: Q6K, Size: 6254198880, Files: []string{"qwen2.5-7b-instruct-q6_k-00001-of-00002.gguf", "qwen2.5-7b-instruct-q6_k-00002-of-00002.gguf"}},
			{Quant: Q8_0, Size: 8098525408, Files: []string{"qwen2.5-7b-instruct-q8_0-00001-of-00003.gguf", "qwen2.5-7b-instruct-q8_0-00002-of-00003.gguf", "qwen2.5-7b-instruct-q8_0-00003-of-00003.gguf"}},
			{Quant: FP16, Size: 15237853536, Files: []string{"qwen2.5-7b-instruct-fp16-00001-of-00004.gguf", "qwen2.5-7b-instruct-fp16-00002-of-00004.gguf", "qwen2.5-7b-instruct-fp16-00003-of-00004.gguf", "qwen2.5-7b-instruct-fp16-00004-of-00004.gguf"}},
		},
	},
	{
		ID: Qwen25_14B, Family: "Qwen2.5", Parameters: "14B", Repo: "Qwen/Qwen2.5-14B-Instruct-GGUF",
		Thinking: false, Description: "Qwen2.5 14B — универсальная чат-модель",
		Variants: []Variant{
			{Quant: Q2K, Size: 5770497600, Files: []string{"qwen2.5-14b-instruct-q2_k-00001-of-00002.gguf", "qwen2.5-14b-instruct-q2_k-00002-of-00002.gguf"}},
			{Quant: Q3KM, Size: 7339204160, Files: []string{"qwen2.5-14b-instruct-q3_k_m-00001-of-00002.gguf", "qwen2.5-14b-instruct-q3_k_m-00002-of-00002.gguf"}},
			{Quant: Q4_0, Size: 8517725856, Files: []string{"qwen2.5-14b-instruct-q4_0-00001-of-00003.gguf", "qwen2.5-14b-instruct-q4_0-00002-of-00003.gguf", "qwen2.5-14b-instruct-q4_0-00003-of-00003.gguf"}},
			{Quant: Q4KM, Size: 8988110496, Files: []string{"qwen2.5-14b-instruct-q4_k_m-00001-of-00003.gguf", "qwen2.5-14b-instruct-q4_k_m-00002-of-00003.gguf", "qwen2.5-14b-instruct-q4_k_m-00003-of-00003.gguf"}},
			{Quant: Q5_0, Size: 10266554048, Files: []string{"qwen2.5-14b-instruct-q5_0-00001-of-00003.gguf", "qwen2.5-14b-instruct-q5_0-00002-of-00003.gguf", "qwen2.5-14b-instruct-q5_0-00003-of-00003.gguf"}},
			{Quant: Q5KM, Size: 10508873376, Files: []string{"qwen2.5-14b-instruct-q5_k_m-00001-of-00003.gguf", "qwen2.5-14b-instruct-q5_k_m-00002-of-00003.gguf", "qwen2.5-14b-instruct-q5_k_m-00003-of-00003.gguf"}},
			{Quant: Q6K, Size: 12124684064, Files: []string{"qwen2.5-14b-instruct-q6_k-00001-of-00004.gguf", "qwen2.5-14b-instruct-q6_k-00002-of-00004.gguf", "qwen2.5-14b-instruct-q6_k-00003-of-00004.gguf", "qwen2.5-14b-instruct-q6_k-00004-of-00004.gguf"}},
			{Quant: Q8_0, Size: 15701597984, Files: []string{"qwen2.5-14b-instruct-q8_0-00001-of-00004.gguf", "qwen2.5-14b-instruct-q8_0-00002-of-00004.gguf", "qwen2.5-14b-instruct-q8_0-00003-of-00004.gguf", "qwen2.5-14b-instruct-q8_0-00004-of-00004.gguf"}},
			{Quant: FP16, Size: 29547716864, Files: []string{"qwen2.5-14b-instruct-fp16-00001-of-00008.gguf", "qwen2.5-14b-instruct-fp16-00002-of-00008.gguf", "qwen2.5-14b-instruct-fp16-00003-of-00008.gguf", "qwen2.5-14b-instruct-fp16-00004-of-00008.gguf", "qwen2.5-14b-instruct-fp16-00005-of-00008.gguf", "qwen2.5-14b-instruct-fp16-00006-of-00008.gguf", "qwen2.5-14b-instruct-fp16-00007-of-00008.gguf", "qwen2.5-14b-instruct-fp16-00008-of-00008.gguf"}},
		},
	},
	{
		ID: Qwen25_32B, Family: "Qwen2.5", Parameters: "32B", Repo: "Qwen/Qwen2.5-32B-Instruct-GGUF",
		Thinking: false, Description: "Qwen2.5 32B — универсальная чат-модель",
		Variants: []Variant{
			{Quant: Q2K, Size: 12313098784, Files: []string{"qwen2.5-32b-instruct-q2_k-00001-of-00004.gguf", "qwen2.5-32b-instruct-q2_k-00002-of-00004.gguf", "qwen2.5-32b-instruct-q2_k-00003-of-00004.gguf", "qwen2.5-32b-instruct-q2_k-00004-of-00004.gguf"}},
			{Quant: Q3KM, Size: 15935048320, Files: []string{"qwen2.5-32b-instruct-q3_k_m-00001-of-00005.gguf", "qwen2.5-32b-instruct-q3_k_m-00002-of-00005.gguf", "qwen2.5-32b-instruct-q3_k_m-00003-of-00005.gguf", "qwen2.5-32b-instruct-q3_k_m-00004-of-00005.gguf", "qwen2.5-32b-instruct-q3_k_m-00005-of-00005.gguf"}},
			{Quant: Q4_0, Size: 18640231072, Files: []string{"qwen2.5-32b-instruct-q4_0-00001-of-00005.gguf", "qwen2.5-32b-instruct-q4_0-00002-of-00005.gguf", "qwen2.5-32b-instruct-q4_0-00003-of-00005.gguf", "qwen2.5-32b-instruct-q4_0-00004-of-00005.gguf", "qwen2.5-32b-instruct-q4_0-00005-of-00005.gguf"}},
			{Quant: Q4KM, Size: 19851336384, Files: []string{"qwen2.5-32b-instruct-q4_k_m-00001-of-00005.gguf", "qwen2.5-32b-instruct-q4_k_m-00002-of-00005.gguf", "qwen2.5-32b-instruct-q4_k_m-00003-of-00005.gguf", "qwen2.5-32b-instruct-q4_k_m-00004-of-00005.gguf", "qwen2.5-32b-instruct-q4_k_m-00005-of-00005.gguf"}},
			{Quant: Q5_0, Size: 22638254880, Files: []string{"qwen2.5-32b-instruct-q5_0-00001-of-00006.gguf", "qwen2.5-32b-instruct-q5_0-00002-of-00006.gguf", "qwen2.5-32b-instruct-q5_0-00003-of-00006.gguf", "qwen2.5-32b-instruct-q5_0-00004-of-00006.gguf", "qwen2.5-32b-instruct-q5_0-00005-of-00006.gguf", "qwen2.5-32b-instruct-q5_0-00006-of-00006.gguf"}},
			{Quant: Q5KM, Size: 23262157568, Files: []string{"qwen2.5-32b-instruct-q5_k_m-00001-of-00006.gguf", "qwen2.5-32b-instruct-q5_k_m-00002-of-00006.gguf", "qwen2.5-32b-instruct-q5_k_m-00003-of-00006.gguf", "qwen2.5-32b-instruct-q5_k_m-00004-of-00006.gguf", "qwen2.5-32b-instruct-q5_k_m-00005-of-00006.gguf", "qwen2.5-32b-instruct-q5_k_m-00006-of-00006.gguf"}},
			{Quant: Q6K, Size: 26886155104, Files: []string{"qwen2.5-32b-instruct-q6_k-00001-of-00007.gguf", "qwen2.5-32b-instruct-q6_k-00002-of-00007.gguf", "qwen2.5-32b-instruct-q6_k-00003-of-00007.gguf", "qwen2.5-32b-instruct-q6_k-00004-of-00007.gguf", "qwen2.5-32b-instruct-q6_k-00005-of-00007.gguf", "qwen2.5-32b-instruct-q6_k-00006-of-00007.gguf", "qwen2.5-32b-instruct-q6_k-00007-of-00007.gguf"}},
			{Quant: Q8_0, Size: 34820885632, Files: []string{"qwen2.5-32b-instruct-q8_0-00001-of-00009.gguf", "qwen2.5-32b-instruct-q8_0-00002-of-00009.gguf", "qwen2.5-32b-instruct-q8_0-00003-of-00009.gguf", "qwen2.5-32b-instruct-q8_0-00004-of-00009.gguf", "qwen2.5-32b-instruct-q8_0-00005-of-00009.gguf", "qwen2.5-32b-instruct-q8_0-00006-of-00009.gguf", "qwen2.5-32b-instruct-q8_0-00007-of-00009.gguf", "qwen2.5-32b-instruct-q8_0-00008-of-00009.gguf", "qwen2.5-32b-instruct-q8_0-00009-of-00009.gguf"}},
			{Quant: FP16, Size: 65535971488, Files: []string{"qwen2.5-32b-instruct-fp16-00001-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00002-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00003-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00004-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00005-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00006-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00007-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00008-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00009-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00010-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00011-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00012-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00013-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00014-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00015-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00016-of-00017.gguf", "qwen2.5-32b-instruct-fp16-00017-of-00017.gguf"}},
		},
	},
	{
		ID: Qwen25_72B, Family: "Qwen2.5", Parameters: "72B", Repo: "Qwen/Qwen2.5-72B-Instruct-GGUF",
		Thinking: false, Description: "Qwen2.5 72B — универсальная чат-модель",
		Variants: []Variant{
			{Quant: Q2K, Size: 27329259168, Files: []string{"qwen2.5-72b-instruct-q2_k-00001-of-00007.gguf", "qwen2.5-72b-instruct-q2_k-00002-of-00007.gguf", "qwen2.5-72b-instruct-q2_k-00003-of-00007.gguf", "qwen2.5-72b-instruct-q2_k-00004-of-00007.gguf", "qwen2.5-72b-instruct-q2_k-00005-of-00007.gguf", "qwen2.5-72b-instruct-q2_k-00006-of-00007.gguf", "qwen2.5-72b-instruct-q2_k-00007-of-00007.gguf"}},
			{Quant: Q3KM, Size: 35472140224, Files: []string{"qwen2.5-72b-instruct-q3_k_m-00001-of-00009.gguf", "qwen2.5-72b-instruct-q3_k_m-00002-of-00009.gguf", "qwen2.5-72b-instruct-q3_k_m-00003-of-00009.gguf", "qwen2.5-72b-instruct-q3_k_m-00004-of-00009.gguf", "qwen2.5-72b-instruct-q3_k_m-00005-of-00009.gguf", "qwen2.5-72b-instruct-q3_k_m-00006-of-00009.gguf", "qwen2.5-72b-instruct-q3_k_m-00007-of-00009.gguf", "qwen2.5-72b-instruct-q3_k_m-00008-of-00009.gguf", "qwen2.5-72b-instruct-q3_k_m-00009-of-00009.gguf"}},
			{Quant: Q4_0, Size: 41373296800, Files: []string{"qwen2.5-72b-instruct-q4_0-00001-of-00011.gguf", "qwen2.5-72b-instruct-q4_0-00002-of-00011.gguf", "qwen2.5-72b-instruct-q4_0-00003-of-00011.gguf", "qwen2.5-72b-instruct-q4_0-00004-of-00011.gguf", "qwen2.5-72b-instruct-q4_0-00005-of-00011.gguf", "qwen2.5-72b-instruct-q4_0-00006-of-00011.gguf", "qwen2.5-72b-instruct-q4_0-00007-of-00011.gguf", "qwen2.5-72b-instruct-q4_0-00008-of-00011.gguf", "qwen2.5-72b-instruct-q4_0-00009-of-00011.gguf", "qwen2.5-72b-instruct-q4_0-00010-of-00011.gguf", "qwen2.5-72b-instruct-q4_0-00011-of-00011.gguf"}},
			{Quant: Q4KM, Size: 44010465472, Files: []string{"qwen2.5-72b-instruct-q4_k_m-00001-of-00012.gguf", "qwen2.5-72b-instruct-q4_k_m-00002-of-00012.gguf", "qwen2.5-72b-instruct-q4_k_m-00003-of-00012.gguf", "qwen2.5-72b-instruct-q4_k_m-00004-of-00012.gguf", "qwen2.5-72b-instruct-q4_k_m-00005-of-00012.gguf", "qwen2.5-72b-instruct-q4_k_m-00006-of-00012.gguf", "qwen2.5-72b-instruct-q4_k_m-00007-of-00012.gguf", "qwen2.5-72b-instruct-q4_k_m-00008-of-00012.gguf", "qwen2.5-72b-instruct-q4_k_m-00009-of-00012.gguf", "qwen2.5-72b-instruct-q4_k_m-00010-of-00012.gguf", "qwen2.5-72b-instruct-q4_k_m-00011-of-00012.gguf", "qwen2.5-72b-instruct-q4_k_m-00012-of-00012.gguf"}},
			{Quant: Q5_0, Size: 50337049024, Files: []string{"qwen2.5-72b-instruct-q5_0-00001-of-00013.gguf", "qwen2.5-72b-instruct-q5_0-00002-of-00013.gguf", "qwen2.5-72b-instruct-q5_0-00003-of-00013.gguf", "qwen2.5-72b-instruct-q5_0-00004-of-00013.gguf", "qwen2.5-72b-instruct-q5_0-00005-of-00013.gguf", "qwen2.5-72b-instruct-q5_0-00006-of-00013.gguf", "qwen2.5-72b-instruct-q5_0-00007-of-00013.gguf", "qwen2.5-72b-instruct-q5_0-00008-of-00013.gguf", "qwen2.5-72b-instruct-q5_0-00009-of-00013.gguf", "qwen2.5-72b-instruct-q5_0-00010-of-00013.gguf", "qwen2.5-72b-instruct-q5_0-00011-of-00013.gguf", "qwen2.5-72b-instruct-q5_0-00012-of-00013.gguf", "qwen2.5-72b-instruct-q5_0-00013-of-00013.gguf"}},
			{Quant: Q5KM, Size: 51673983488, Files: []string{"qwen2.5-72b-instruct-q5_k_m-00001-of-00014.gguf", "qwen2.5-72b-instruct-q5_k_m-00002-of-00014.gguf", "qwen2.5-72b-instruct-q5_k_m-00003-of-00014.gguf", "qwen2.5-72b-instruct-q5_k_m-00004-of-00014.gguf", "qwen2.5-72b-instruct-q5_k_m-00005-of-00014.gguf", "qwen2.5-72b-instruct-q5_k_m-00006-of-00014.gguf", "qwen2.5-72b-instruct-q5_k_m-00007-of-00014.gguf", "qwen2.5-72b-instruct-q5_k_m-00008-of-00014.gguf", "qwen2.5-72b-instruct-q5_k_m-00009-of-00014.gguf", "qwen2.5-72b-instruct-q5_k_m-00010-of-00014.gguf", "qwen2.5-72b-instruct-q5_k_m-00011-of-00014.gguf", "qwen2.5-72b-instruct-q5_k_m-00012-of-00014.gguf", "qwen2.5-72b-instruct-q5_k_m-00013-of-00014.gguf", "qwen2.5-72b-instruct-q5_k_m-00014-of-00014.gguf"}},
			{Quant: Q6K, Size: 59861035808, Files: []string{"qwen2.5-72b-instruct-q6_k-00001-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00002-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00003-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00004-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00005-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00006-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00007-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00008-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00009-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00010-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00011-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00012-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00013-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00014-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00015-of-00016.gguf", "qwen2.5-72b-instruct-q6_k-00016-of-00016.gguf"}},
			{Quant: Q8_0, Size: 77530000800, Files: []string{"qwen2.5-72b-instruct-q8_0-00001-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00002-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00003-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00004-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00005-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00006-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00007-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00008-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00009-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00010-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00011-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00012-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00013-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00014-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00015-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00016-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00017-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00018-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00019-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00020-of-00021.gguf", "qwen2.5-72b-instruct-q8_0-00021-of-00021.gguf"}},
			{Quant: FP16, Size: 145925994464, Files: []string{"qwen2.5-72b-instruct-fp16-00001-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00002-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00003-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00004-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00005-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00006-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00007-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00008-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00009-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00010-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00011-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00012-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00013-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00014-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00015-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00016-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00017-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00018-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00019-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00020-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00021-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00022-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00023-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00024-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00025-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00026-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00027-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00028-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00029-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00030-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00031-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00032-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00033-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00034-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00035-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00036-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00037-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00038-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00039-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00040-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00041-of-00042.gguf", "qwen2.5-72b-instruct-fp16-00042-of-00042.gguf"}},
		},
	},
	{
		ID: Qwen25Coder_05B, Family: "Qwen2.5-Coder", Parameters: "0.5B", Repo: "Qwen/Qwen2.5-Coder-0.5B-Instruct-GGUF",
		Thinking: false, Description: "Qwen2.5-Coder 0.5B — модель для программирования",
		Variants: []Variant{
			{Quant: Q2K, Size: 415182720, Files: []string{"qwen2.5-coder-0.5b-instruct-q2_k.gguf"}},
			{Quant: Q3KM, Size: 432041856, Files: []string{"qwen2.5-coder-0.5b-instruct-q3_k_m.gguf"}},
			{Quant: Q4_0, Size: 428730240, Files: []string{"qwen2.5-coder-0.5b-instruct-q4_0.gguf"}},
			{Quant: Q4KM, Size: 491400064, Files: []string{"qwen2.5-coder-0.5b-instruct-q4_k_m.gguf"}},
			{Quant: Q5_0, Size: 490475392, Files: []string{"qwen2.5-coder-0.5b-instruct-q5_0.gguf"}},
			{Quant: Q5KM, Size: 522186624, Files: []string{"qwen2.5-coder-0.5b-instruct-q5_k_m.gguf"}},
			{Quant: Q6K, Size: 650379136, Files: []string{"qwen2.5-coder-0.5b-instruct-q6_k.gguf"}},
			{Quant: Q8_0, Size: 675710848, Files: []string{"qwen2.5-coder-0.5b-instruct-q8_0.gguf"}},
			{Quant: FP16, Size: 1266425728, Files: []string{"qwen2.5-coder-0.5b-instruct-fp16.gguf"}},
		},
	},
	{
		ID: Qwen25Coder_15B, Family: "Qwen2.5-Coder", Parameters: "1.5B", Repo: "Qwen/Qwen2.5-Coder-1.5B-Instruct-GGUF",
		Thinking: false, Description: "Qwen2.5-Coder 1.5B — модель для программирования",
		Variants: []Variant{
			{Quant: Q2K, Size: 752880192, Files: []string{"qwen2.5-coder-1.5b-instruct-q2_k.gguf"}},
			{Quant: Q3KM, Size: 924456000, Files: []string{"qwen2.5-coder-1.5b-instruct-q3_k_m.gguf"}},
			{Quant: Q4_0, Size: 1066227264, Files: []string{"qwen2.5-coder-1.5b-instruct-q4_0.gguf"}},
			{Quant: Q4KM, Size: 1117320768, Files: []string{"qwen2.5-coder-1.5b-instruct-q4_k_m.gguf"}},
			{Quant: Q5_0, Size: 1259173440, Files: []string{"qwen2.5-coder-1.5b-instruct-q5_0.gguf"}},
			{Quant: Q5KM, Size: 1285494336, Files: []string{"qwen2.5-coder-1.5b-instruct-q5_k_m.gguf"}},
			{Quant: Q6K, Size: 1464178752, Files: []string{"qwen2.5-coder-1.5b-instruct-q6_k.gguf"}},
			{Quant: Q8_0, Size: 1894532160, Files: []string{"qwen2.5-coder-1.5b-instruct-q8_0.gguf"}},
		},
	},
	{
		ID: Qwen25Coder_3B, Family: "Qwen2.5-Coder", Parameters: "3B", Repo: "Qwen/Qwen2.5-Coder-3B-Instruct-GGUF",
		Thinking: false, Description: "Qwen2.5-Coder 3B — модель для программирования",
		Variants: []Variant{
			{Quant: Q2K, Size: 1376856512, Files: []string{"qwen2.5-coder-3b-instruct-q2_k.gguf"}},
			{Quant: Q3KM, Size: 1724178880, Files: []string{"qwen2.5-coder-3b-instruct-q3_k_m.gguf"}},
			{Quant: Q4_0, Size: 1997879744, Files: []string{"qwen2.5-coder-3b-instruct-q4_0.gguf"}},
			{Quant: Q4KM, Size: 2104932800, Files: []string{"qwen2.5-coder-3b-instruct-q4_k_m.gguf"}},
			{Quant: Q5_0, Size: 2383591872, Files: []string{"qwen2.5-coder-3b-instruct-q5_0.gguf"}},
			{Quant: Q5KM, Size: 2438740416, Files: []string{"qwen2.5-coder-3b-instruct-q5_k_m.gguf"}},
			{Quant: Q6K, Size: 2793411008, Files: []string{"qwen2.5-coder-3b-instruct-q6_k.gguf"}},
			{Quant: Q8_0, Size: 3616088512, Files: []string{"qwen2.5-coder-3b-instruct-q8_0.gguf"}},
			{Quant: FP16, Size: 6800646592, Files: []string{"qwen2.5-coder-3b-instruct-fp16.gguf"}},
		},
	},
	{
		ID: Qwen25Coder_7B, Family: "Qwen2.5-Coder", Parameters: "7B", Repo: "Qwen/Qwen2.5-Coder-7B-Instruct-GGUF",
		Thinking: false, Description: "Qwen2.5-Coder 7B — модель для программирования",
		Variants: []Variant{
			{Quant: Q2K, Size: 3015940032, Files: []string{"qwen2.5-coder-7b-instruct-q2_k.gguf"}},
			{Quant: Q3KM, Size: 3808391104, Files: []string{"qwen2.5-coder-7b-instruct-q3_k_m.gguf"}},
			{Quant: Q4_0, Size: 4431390720, Files: []string{"qwen2.5-coder-7b-instruct-q4_0.gguf"}},
			{Quant: Q4KM, Size: 4683073536, Files: []string{"qwen2.5-coder-7b-instruct-q4_k_m.gguf"}},
			{Quant: Q5_0, Size: 5315176576, Files: []string{"qwen2.5-coder-7b-instruct-q5_0-00001-of-00002.gguf", "qwen2.5-coder-7b-instruct-q5_0-00002-of-00002.gguf"}},
			{Quant: Q5KM, Size: 5444831232, Files: []string{"qwen2.5-coder-7b-instruct-q5_k_m.gguf"}},
			{Quant: Q6K, Size: 6254198784, Files: []string{"qwen2.5-coder-7b-instruct-q6_k.gguf"}},
			{Quant: Q8_0, Size: 8098525184, Files: []string{"qwen2.5-coder-7b-instruct-q8_0.gguf"}},
			{Quant: FP16, Size: 15237853184, Files: []string{"qwen2.5-coder-7b-instruct-fp16.gguf"}},
		},
	},
	{
		ID: Qwen25Coder_14B, Family: "Qwen2.5-Coder", Parameters: "14B", Repo: "Qwen/Qwen2.5-Coder-14B-Instruct-GGUF",
		Thinking: false, Description: "Qwen2.5-Coder 14B — модель для программирования",
		Variants: []Variant{
			{Quant: Q2K, Size: 5770497408, Files: []string{"qwen2.5-coder-14b-instruct-q2_k.gguf"}},
			{Quant: Q3KM, Size: 7339203968, Files: []string{"qwen2.5-coder-14b-instruct-q3_k_m.gguf"}},
			{Quant: Q4_0, Size: 8517725568, Files: []string{"qwen2.5-coder-14b-instruct-q4_0.gguf"}},
			{Quant: Q4KM, Size: 8988110272, Files: []string{"qwen2.5-coder-14b-instruct-q4_k_m.gguf"}},
			{Quant: Q5_0, Size: 10266553792, Files: []string{"qwen2.5-coder-14b-instruct-q5_0.gguf"}},
			{Quant: Q5KM, Size: 10508873152, Files: []string{"qwen2.5-coder-14b-instruct-q5_k_m.gguf"}},
			{Quant: Q6K, Size: 12124683712, Files: []string{"qwen2.5-coder-14b-instruct-q6_k.gguf"}},
			{Quant: Q8_0, Size: 15701597632, Files: []string{"qwen2.5-coder-14b-instruct-q8_0.gguf"}},
			{Quant: FP16, Size: 29547716032, Files: []string{"qwen2.5-coder-14b-instruct-fp16.gguf"}},
		},
	},
	{
		ID: Qwen25Coder_32B, Family: "Qwen2.5-Coder", Parameters: "32B", Repo: "Qwen/Qwen2.5-Coder-32B-Instruct-GGUF",
		Thinking: false, Description: "Qwen2.5-Coder 32B — модель для программирования",
		Variants: []Variant{
			{Quant: Q2K, Size: 12313098432, Files: []string{"qwen2.5-coder-32b-instruct-q2_k.gguf"}},
			{Quant: Q3KM, Size: 15935047872, Files: []string{"qwen2.5-coder-32b-instruct-q3_k_m.gguf"}},
			{Quant: Q4_0, Size: 18640230592, Files: []string{"qwen2.5-coder-32b-instruct-q4_0.gguf"}},
			{Quant: Q4KM, Size: 19851335872, Files: []string{"qwen2.5-coder-32b-instruct-q4_k_m.gguf"}},
			{Quant: Q5_0, Size: 22638254272, Files: []string{"qwen2.5-coder-32b-instruct-q5_0.gguf"}},
			{Quant: Q5KM, Size: 23262156992, Files: []string{"qwen2.5-coder-32b-instruct-q5_k_m.gguf"}},
			{Quant: Q6K, Size: 26886154432, Files: []string{"qwen2.5-coder-32b-instruct-q6_k.gguf"}},
			{Quant: Q8_0, Size: 34820884672, Files: []string{"qwen2.5-coder-32b-instruct-q8_0.gguf"}},
			{Quant: FP16, Size: 65535970464, Files: []string{"qwen2.5-coder-32b-instruct-fp16-00001-of-00009.gguf", "qwen2.5-coder-32b-instruct-fp16-00002-of-00009.gguf", "qwen2.5-coder-32b-instruct-fp16-00003-of-00009.gguf", "qwen2.5-coder-32b-instruct-fp16-00004-of-00009.gguf", "qwen2.5-coder-32b-instruct-fp16-00005-of-00009.gguf", "qwen2.5-coder-32b-instruct-fp16-00006-of-00009.gguf", "qwen2.5-coder-32b-instruct-fp16-00007-of-00009.gguf", "qwen2.5-coder-32b-instruct-fp16-00008-of-00009.gguf", "qwen2.5-coder-32b-instruct-fp16-00009-of-00009.gguf"}},
		},
	},
	{
		ID: Qwen3_06B, Family: "Qwen3", Parameters: "0.6B", Repo: "Qwen/Qwen3-0.6B-GGUF",
		Thinking: true, Description: "Qwen3 0.6B — чат-модель с режимом рассуждений",
		Variants: []Variant{
			{Quant: Q8_0, Size: 639446688, Files: []string{"Qwen3-0.6B-Q8_0.gguf"}},
		},
	},
	{
		ID: Qwen3_17B, Family: "Qwen3", Parameters: "1.7B", Repo: "Qwen/Qwen3-1.7B-GGUF",
		Thinking: true, Description: "Qwen3 1.7B — чат-модель с режимом рассуждений",
		Variants: []Variant{
			{Quant: Q8_0, Size: 1834426016, Files: []string{"Qwen3-1.7B-Q8_0.gguf"}},
		},
	},
	{
		ID: Qwen3_4B, Family: "Qwen3", Parameters: "4B", Repo: "Qwen/Qwen3-4B-GGUF",
		Thinking: true, Description: "Qwen3 4B — чат-модель с режимом рассуждений",
		Variants: []Variant{
			{Quant: Q4KM, Size: 2497280256, Files: []string{"Qwen3-4B-Q4_K_M.gguf"}},
			{Quant: Q5_0, Size: 2823710976, Files: []string{"Qwen3-4B-Q5_0.gguf"}},
			{Quant: Q5KM, Size: 2889513184, Files: []string{"Qwen3-4B-Q5_K_M.gguf"}},
			{Quant: Q6K, Size: 3306260704, Files: []string{"Qwen3-4B-Q6_K.gguf"}},
			{Quant: Q8_0, Size: 4280404704, Files: []string{"Qwen3-4B-Q8_0.gguf"}},
		},
	},
	{
		ID: Qwen3_8B, Family: "Qwen3", Parameters: "8B", Repo: "Qwen/Qwen3-8B-GGUF",
		Thinking: true, Description: "Qwen3 8B — чат-модель с режимом рассуждений",
		Variants: []Variant{
			{Quant: Q4KM, Size: 5027783488, Files: []string{"Qwen3-8B-Q4_K_M.gguf"}},
			{Quant: Q5_0, Size: 5720761152, Files: []string{"Qwen3-8B-Q5_0.gguf"}},
			{Quant: Q5KM, Size: 5851112224, Files: []string{"Qwen3-8B-Q5_K_M.gguf"}},
			{Quant: Q6K, Size: 6725899040, Files: []string{"Qwen3-8B-Q6_K.gguf"}},
			{Quant: Q8_0, Size: 8709518112, Files: []string{"Qwen3-8B-Q8_0.gguf"}},
		},
	},
	{
		ID: Qwen3_14B, Family: "Qwen3", Parameters: "14B", Repo: "Qwen/Qwen3-14B-GGUF",
		Thinking: true, Description: "Qwen3 14B — чат-модель с режимом рассуждений",
		Variants: []Variant{
			{Quant: Q4KM, Size: 9001752960, Files: []string{"Qwen3-14B-Q4_K_M.gguf"}},
			{Quant: Q5_0, Size: 10263894400, Files: []string{"Qwen3-14B-Q5_0.gguf"}},
			{Quant: Q5KM, Size: 10514569568, Files: []string{"Qwen3-14B-Q5_K_M.gguf"}},
			{Quant: Q6K, Size: 12121937248, Files: []string{"Qwen3-14B-Q6_K.gguf"}},
			{Quant: Q8_0, Size: 15698533728, Files: []string{"Qwen3-14B-Q8_0.gguf"}},
		},
	},
	{
		ID: Qwen3_32B, Family: "Qwen3", Parameters: "32B", Repo: "Qwen/Qwen3-32B-GGUF",
		Thinking: true, Description: "Qwen3 32B — чат-модель с режимом рассуждений",
		Variants: []Variant{
			{Quant: Q4KM, Size: 19762149024, Files: []string{"Qwen3-32B-Q4_K_M.gguf"}},
			{Quant: Q5_0, Size: 22635493024, Files: []string{"Qwen3-32B-Q5_0.gguf"}},
			{Quant: Q5KM, Size: 23214831232, Files: []string{"Qwen3-32B-Q5_K_M.gguf"}},
			{Quant: Q6K, Size: 26883306112, Files: []string{"Qwen3-32B-Q6_K.gguf"}},
			{Quant: Q8_0, Size: 34817718912, Files: []string{"Qwen3-32B-Q8_0.gguf"}},
		},
	},
}
