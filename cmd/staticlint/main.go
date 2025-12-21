// Package main — staticlint: набор статических анализаторов, объединённых через multichecker.
//
// staticlint формирует срез *analysis.Analyzer и передаёт его в multichecker.Main
// для запуска над пакетами.
//
// Включает:
//   - printf — проверяет вызовы printf-подобных функций: соответствие формата и
//     переданных аргументов, предотвращает паники и неправильный вывод.
//   - shadow — обнаруживает затемнение переменных (variable shadowing), помогает
//     избежать скрытых багов из‑за повторного объявления переменных.
//   - structtag — валидирует теги в полях структур (формат `json`, `xml` и т.д.),
//     ловит опечатки и некорректные теги.
//   - staticcheck (SA*) — набор статических проверок уровня SA, выявляет серьёзные
//     ошибки и потенциальные дефекты (ненужные операции, утечки, неверное
//     использование API и т.п.).
//   - staticcheck ST1013 — рекомендует использовать константы из пакета net/http
//     вместо жёсткого кодирования числовых кодов состояния HTTP.
//   - ineffassign — находит неэффективные присваивания (когда значение записывается
//     в переменную, но затем не используется), помогает устранить мёртвый код и
//     логические ошибки.
//   - exhaustive — проверяет конструкции switch на исчерпывающую обработку всех
//     значений перечислений/констант.
//   - osexitcheck — пользовательский анализатор, запрещающий использовать прямой
//     вызов os.Exit в функции main пакета main.
//
// Для запуска staticlint необходимо собрать бинарник:
//
//	go build -o staticlint ./cmd/staticlint
//
// Затем выполнить анализ кода проекта:
//
//	./staticlint ./...
package main

import (
	"github.com/Skifskii/link-shortener/cmd/staticlint/osexitcheck"
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/nishanths/exhaustive"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

// main собирает анализаторы в срез checks и вызывает multichecker.Main(checks...).
func main() {
	var checks []*analysis.Analyzer

	// Добавляем стандартные анализаторы из analysis/passes
	checks = append(checks, printf.Analyzer)
	checks = append(checks, shadow.Analyzer)
	checks = append(checks, structtag.Analyzer)

	// Добавляем все проверки staticcheck с префиксом SA
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[:2] == "SA" || v.Analyzer.Name == "ST1013" {
			checks = append(checks, v.Analyzer)
		}
	}

	// Добавляем другие анализаторы
	checks = append(checks, ineffassign.Analyzer)
	checks = append(checks, exhaustive.Analyzer)

	// Добавляем наш пользовательский анализатор
	checks = append(checks, osexitcheck.Analyzer)

	multichecker.Main(checks...)
}
