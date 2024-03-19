Есть функиця, которая что-то там ищет по файлу. Но делает она это не очень быстро. Надо её оптимизировать.

Задание на работу с профайлером pprof.

Цель задания - научиться работать с pprof, находить горячие места в коде, уметь строить профиль потребления cpu и памяти, оптимизировать код с учетом этой информации. Написание самого быстрого решения не является целью задания.

Для генерации графа вам понадобится graphviz. Для пользователей windows не забудьте добавить его в PATH чтобы была доступна команда dot.

Рекомендую внимательно прочитать доп. материалы на русском - там ещё много примеров оптимизации и объяснений как работать с профайлером. Фактически там есть вся информация для выполнения этого задания.

Есть с десяток мест где можно оптимизировать.
Вам надо писать отчет, где вы заоптимайзили и что. Со скриншотами и объяснением что делали. Чтобы именно научиться в pprof находить проблемы, а не прикинуть мозгами и решить что вот тут медленно.

Для выполнения задания необходимо чтобы один из параметров ( ns/op, B/op, allocs/op ) был быстрее чем в *BenchmarkSolution* ( fast < solution ) и ещё один лучше *BenchmarkSolution* + 20% ( fast < solution * 1.2), например ( fast allocs/op < 10422*1.2=12506 ).

По памяти ( B/op ) и количеству аллокаций ( allocs/op ) можно ориентироваться ровно на результаты *BenchmarkSolution* ниже, по времени ( ns/op ) - нет, зависит от системы.

Параллелить (использовать горутины) или sync.Pool в это задании не нужно.

Результат в fast.go в функцию FastSearch (изначально там то же самое что в SlowSearch).

Пример результатов с которыми будет сравниваться:
```
$ go test -bench . -benchmem

goos: windows

goarch: amd64

BenchmarkSlow-8 10 142703250 ns/op 336887900 B/op 284175 allocs/op

BenchmarkSolution-8 500 2782432 ns/op 559910 B/op 10422 allocs/op

PASS

ok coursera/hw3 3.897s
```

Запуск:
* `go test -v` - чтобы проверить что ничего не сломалось
* `go test -bench . -benchmem` - для просмотра производительности
* `go tool pprof -http=:8083 /path/ho/bin /path/to/out` - веб-интерфейс для pprof, пользуйтесь им для поиска горячих мест. Не забывайте, что у вас 2 режиме - cpu и mem, там разные out-файлы.

Советы:
* Смотрите где мы аллоцируем память
* Смотрите где мы накапливаем весь результат, хотя нам все значения одновременно не нужны
* Смотрите где происходят преобразования типов, которые можно избежать
* Смотрите не только на графе, но и в pprof в текстовом виде (list FastSearch) - там прямо по исходнику можно увидеть где что
* Задание предполагает использование easyjson. На сервере эта библиотека есть, подключать можно. Но сгенерированный через easyjson код вам надо поместить в файл с вашей функцией
* Можно сделать без easyjson

Примечание:
* easyjson основан на рефлекции и не может работать с пакетом main. Для генерации кода вам необходимо вынести вашу структуру в отдельный пакет, сгенерить там код, потом забрать его в main

---
регулярку заранее создать
контейнз лучше регулярки
капасити слайса лучше задавать так чтобы он не расширялся
кодогенерация быстрее рефлексии
поточная обработка данных
---
go test -bench . -benchmem -cpuprofile=cpu.out -memprofile=mem.out -memprofilerate=1 . 

go tool pprof -http=:8083 hw3.test.exe cpu.out

go tool pprof hw3.test.exe cpu.out
(pprof) top
Showing top 10 nodes out of 153
      flat  flat%   sum%        cum   cum%
     130ms 22.81% 22.81%      130ms 22.81%  runtime.addspecial
      70ms 12.28% 35.09%       80ms 14.04%  runtime.step
      60ms 10.53% 45.61%      170ms 29.82%  runtime.pcvalue
      60ms 10.53% 56.14%       60ms 10.53%  runtime.stdcall3
      30ms  5.26% 61.40%       30ms  5.26%  runtime.funcInfo.entry
      20ms  3.51% 64.91%       20ms  3.51%  runtime.findfunc
      20ms  3.51% 68.42%      220ms 38.60%  runtime.gentraceback
      10ms  1.75% 70.18%      120ms 21.05%  regexp/syntax.(*parser).push
      10ms  1.75% 71.93%       10ms  1.75%  runtime.(*randomEnum).next
      10ms  1.75% 73.68%       10ms  1.75%  runtime.(*scavengerState).run

(pprof) list SlowSearch
 130ms     38:           err := json.Unmarshal([]byte(line), &user)
100ms     62:                   if ok, err := regexp.MatchString("Android", browser); ok && err == nil {
 170ms     84:                   if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {
     10ms    106:           foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)

go tool pprof hw3.test.exe mem.out
(pprof) alloc_space
(pprof) top 
      flat  flat%   sum%        cum   cum%
 7255.25kB 28.64% 28.64%  7255.25kB 28.64%  regexp/syntax.(*compiler).inst (inline)
 5498.25kB 21.70% 50.35%  5498.25kB 21.70%  io.ReadAll
 2626.20kB 10.37% 60.71%  2626.20kB 10.37%  regexp/syntax.(*parser).newRegexp (inline)
 1974.24kB  7.79% 68.51% 22622.22kB 89.30%  hw3.SlowSearch
 1303.06kB  5.14% 73.65% 13194.76kB 52.09%  regexp.compile
 1153.95kB  4.56% 78.21%  1153.95kB  4.56%  runtime/pprof.StartCPUProfile
 1000.38kB  3.95% 82.16%  4064.09kB 16.04%  regexp/syntax.parse
     648kB  2.56% 84.71%  1171.88kB  4.63%  compress/flate.NewWriter
  409.30kB  1.62% 86.33%   409.30kB  1.62%  encoding/json.unquote (inline)
  375.05kB  1.48% 87.81%   750.14kB  2.96%  regexp/syntax.(*compiler).init (inline)

(pprof) list SlowSearch 
 5.37MB     22:   fileContents, err := ioutil.ReadAll(file)
 1.09MB     1.11MB     32:   lines := strings.Split(string(fileContents), "\n")
 580.66kB     2.10MB     38:           err := json.Unmarshal([]byte(line), &user)
 7.94MB     62:                   if ok, err := regexp.MatchString("Android", browser); ok && err == nil {
 5.27MB     84:                   if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {
 185.53kB   199.23kB    106:           foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)


(pprof) alloc_objects
      flat  flat%   sum%        cum   cum%
     36017 20.32% 20.32%      36017 20.32%  regexp/syntax.(*compiler).inst (inline)
     24011 13.54% 33.86%      24011 13.54%  regexp/syntax.(*parser).newRegexp (inline)
     16001  9.03% 42.89%      28001 15.80%  regexp/syntax.(*parser).push
     12000  6.77% 49.66%      12000  6.77%  regexp/syntax.(*parser).maybeConcat
     11378  6.42% 56.08%     128039 72.23%  regexp.compile
     11078  6.25% 62.32%      11078  6.25%  encoding/json.(*decodeState).literalStore
      8003  4.51% 66.84%      60015 33.85%  regexp/syntax.parse
      8001  4.51% 71.35%       8001  4.51%  regexp/syntax.(*Regexp).CapNames
      8001  4.51% 75.87%      16003  9.03%  regexp/syntax.(*compiler).init (inline)
      8000  4.51% 80.38%       8000  4.51%  reflect.New

(pprof) list SlowSearch
 2000       2000     36:           user := make(map[string]interface{})
   1001      45354     38:           err := json.Unmarshal([]byte(line), &user)
    68043     62:                   if ok, err := regexp.MatchString("Android", browser); ok && err == nil {
           60027     84:                   if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {
             .        363    105:           email := r.ReplaceAllString(user["email"].(string), " [at] ")
       188        308    106:           foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)


