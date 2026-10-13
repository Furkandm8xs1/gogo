# Go Dil Rehberi: Sıfırdan Network Servisine

Bu dosya, projede kullanılan Go sözdizimini ve standart kütüphaneleri sıfırdan
öğrenmek için hazırlanmıştır. Önce temel dili, sonra `main.go` içindeki
`stats` fonksiyonunu ve kullanılan HTTP/concurrency yapılarını açıklıyoruz.

---

## 1. Goroutine nedir: thread mi?

Kısa cevap: **Goroutine, doğrudan bir işletim sistemi thread'i değildir.**

- **Thread (iş parçacığı):** İşletim sisteminin yönettiği çalışma birimidir.
  Oluşturulması ve context switch işlemi görece pahalıdır.
- **Goroutine:** Go runtime tarafından yönetilen, çok daha hafif bir çalışma
  birimidir. Go runtime, binlerce goroutine'i daha az sayıda OS thread'i
  üzerinde çalıştırabilir.

Örneğin:

```go
go server.worker(ctx, 1, &workers)
```

`go` anahtar kelimesi `server.worker` fonksiyonunu yeni bir goroutine içinde
başlatır. `main` fonksiyonu aynı anda çalışmaya devam eder; fonksiyonun
bitmesini beklemez.

Bunu bir restorana benzetebiliriz:

- OS thread'leri: restoranın gerçek çalışanları
- Goroutine'ler: çalışanların sırayla/uygun zamanda yaptığı küçük işler
- Go scheduler: hangi işin hangi çalışanda çalışacağını planlayan sistem

Goroutine'ler thread'lere göre daha ucuzdur; fakat bu, sınırsız sayıda
goroutine açmanın güvenli olduğu anlamına gelmez. Her goroutine kaynak kullanır.
Uzun yaşayan goroutine'lerin nasıl durdurulacağını `context` ile tanımlamak
gerekir.

### Paralellik ve concurrency aynı şey değildir

- **Concurrency (eşzamanlılık):** Birden fazla işin ilerlemesinin yönetilmesi.
- **Parallelism (paralellik):** İşlerin gerçekten aynı anda farklı CPU
  çekirdeklerinde çalışması.

Goroutine'ler concurrency sağlar; Go runtime ve makinenin CPU kapasitesi uygun
olduğunda işler paralel de çalışabilir.

---

## 2. Go programının temel şekli

```go
package main

import "fmt"

func main() {
    fmt.Println("Merhaba Go")
}
```

### `package`

Her Go dosyası bir pakete aittir. Çalıştırılabilir programın paketi
`package main` olmalı ve bu pakette `main()` fonksiyonu bulunmalıdır.

### `import`

Başka paketleri kullanıma alır:

```go
import "fmt"
```

Birden fazla import için proje içinde kullanılan blok biçimi tercih edilir:

```go
import (
    "context"
    "net/http"
    "time"
)
```

Go kullanılmayan import'ları derleme hatası sayar. Bu, kodun temiz kalmasına
yardımcı olur.

### Noktalı virgül

Go çoğu satırın sonuna otomatik olarak noktalı virgül ekler. Bu yüzden genelde
kendimiz `;` yazmayız.

---

## 3. Değişkenler ve sabitler

```go
var port int = 8080
var name = "api"
const maxWorkers = 10
```

Fonksiyon içinde kısa değişken bildirimi de kullanılabilir:

```go
port := 8080
```

`:=` hem değişken oluşturur hem türü değerden çıkarır. Paket seviyesinde
`:=` kullanılamaz; orada `var` veya `const` gerekir.

Birden fazla değer:

```go
name, err := loadConfig()
```

Burada `name` sonucu, `err` ise hata sonucunu alır.

Bir değeri bilerek kullanmıyorsak `_` kullanırız:

```go
func health(w http.ResponseWriter, _ *http.Request) {
    // Request bu handler'da kullanılmıyor.
}
```

`_` blank identifier'dır. Bir değeri atar ve kullanmama niyetimizi belirtir.

---

## 4. Temel türler

Projede kullanılan başlıca türler:

```go
var count int          // tam sayı
var timeout time.Duration
var enabled bool       // true veya false
var message string
```

Diğer yaygın türler:

- `int8`, `int16`, `int32`, `int64`: farklı boyutlarda tam sayılar
- `uint`: negatif olmayan tam sayı
- `float32`, `float64`: ondalıklı sayılar
- `byte`: `uint8` için takma ad
- `rune`: `int32` için takma ad, Unicode karakterleri için kullanılır

Go'da tür dönüşümü açık yazılır:

```go
seconds := int64(5)
```

Go otomatik olarak `int` ile `int64` arasında sessiz dönüşüm yapmaz.

---

## 5. Fonksiyonlar

### Parametresiz fonksiyon

```go
func sayHello() {
    log.Println("hello")
}
```

### Parametreli fonksiyon

```go
func add(a int, b int) int {
    return a + b
}
```

Aynı türdeki parametreler kısaltılabilir:

```go
func add(a, b int) int {
    return a + b
}
```

### Birden fazla dönüş değeri

Go'da bir fonksiyon birden fazla değer döndürebilir. En yaygın kullanım hata
ile sonucunun birlikte dönmesidir:

```go
func parsePort(value string) (int, error) {
    return strconv.Atoi(value)
}

port, err := parsePort("8080")
if err != nil {
    log.Fatal(err)
}
```

Go'da exception yerine genellikle `error` değeri açıkça kontrol edilir:

```go
if err := doSomething(); err != nil {
    return
}
```

Bu yapıda `err` sadece `if` bloğunun kapsamı içinde geçerlidir.

---

## 6. `stats` fonksiyonunu satır satır inceleme

Kodumuz:

```go
func (s *Server) stats(w http.ResponseWriter, _ *http.Request) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    writeJSON(w, http.StatusOK, map[string]int{
        "received":  s.received,
        "processed": s.processed,
        "queued":    len(s.jobs),
    })
}
```

### `func`

Fonksiyon tanımlamaya yarar.

### `(s *Server)` nedir?

Bu bölüm **receiver** olarak adlandırılır. Bu fonksiyonun `Server` türüne ait
bir method olduğunu belirtir.

```go
func (s *Server) stats(...) { ... }
```

Bu method `Server` değerinin `s` adıyla kullanılmasını sağlar. Şu çağrı:

```go
server.stats(w, r)
```

mantıksal olarak şu fikre denktir:

```go
Server.stats(server, w, r)
```

`*Server` pointer receiver'dır. `s`, `Server` nesnesinin kendisi değil,
nesnenin bellekteki adresini taşır. Böylece method aynı `Server` nesnesinin
alanlarına erişebilir ve gerektiğinde onları değiştirebilir.

`s` ismi zorunlu değildir; `server *Server` de yazılabilir. Go'da receiver
isimleri genellikle kısa tutulur.

### `w http.ResponseWriter`

İlk normal parametredir. `http.ResponseWriter`, HTTP cevabını istemciye
yazmak için kullanılan bir interface'tir. Şunları yapabiliriz:

```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
w.Write([]byte(`{"status":"ok"}`))
```

Burada `w` gerçek bir struct olmak zorunda değildir; `ResponseWriter`
interface'ini sağlayan herhangi bir değer olabilir.

### `_ *http.Request`

İkinci parametrenin türü `*http.Request`'tir. Pointer olduğu için başında `*`
vardır. HTTP isteğinin method, URL, header, body ve context bilgilerini taşır.

Bu handler request içeriğini kullanmadığı için parametre adını `_` yaptık.
Normalde şöyle yazılabilir:

```go
func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
    log.Println(r.Method)
}
```

### Neden `stats` başka bir fonksiyon alıyor gibi görünüyor?

Aslında `stats`, fonksiyon almıyor; kendisi iki parametre alan bir HTTP handler
fonksiyonu. Şu satırda:

```go
mux.HandleFunc("GET /stats", s.stats)
```

`s.stats` fonksiyonun çalıştırılması değil, fonksiyonun kendisinin
`HandleFunc`'a verilmesidir. Buna **function value** denir.

`HandleFunc` daha sonra bir HTTP isteği geldiğinde bunu çağırır. Beklediği
fonksiyon şekli şudur:

```go
func(http.ResponseWriter, *http.Request)
```

`s.stats` bu şekle uyar; receiver olan `s` zaten method value oluşturulurken
bağlanmıştır.

Başka bir deyişle:

```go
mux.HandleFunc("GET /stats", s.stats)
```

şu an çalıştırmaz; `/stats` isteği geldiğinde framework'ün çalıştıracağı
callback'i kaydeder.

### `RLock` ve `RUnlock`

`s.mu`, `sync.RWMutex` türündedir:

- `RLock`: veriyi sadece okumak için kilit alır.
- `RUnlock`: okuma kilidini bırakır.
- Birden fazla okuyucu aynı anda okuyabilir.
- Yazma kilidi varken okuyucular bekler.

`stats` aynı anda çalışan HTTP handler'larından `received` ve `processed`
alanlarını okuyor. Worker'lar bu alanları değiştirdiği için kilit kullanmak
gerekir. Aksi halde data race oluşabilir.

### `defer`

```go
s.mu.RLock()
defer s.mu.RUnlock()
```

`defer`, ilgili fonksiyon bitmeden hemen önce verilen çağrıyı çalıştırır.
Fonksiyon normal `return` ile bitse de hata/panic akışı olsa da kilidin
bırakılmasına yardımcı olur.

`defer` satırının fonksiyon çağrısından sonra yazılması önemlidir; böylece
kilit alındıktan sonra bırakılacağı garanti edilir.

### `writeJSON`

```go
writeJSON(w, http.StatusOK, map[string]int{...})
```

Bu, bizim yazdığımız yardımcı fonksiyondur. Üç parametre alır:

1. HTTP cevabının yazılacağı `w`
2. HTTP durum kodu (`200 OK`)
3. JSON'a çevrilecek değer

### `map[string]int`

Map, anahtar-değer koleksiyonudur:

```go
map[string]int{
    "received":  s.received,
    "processed": s.processed,
}
```

Bu map'in anahtarları `string`, değerleri `int` türündedir. JSON'a çevrilince:

```json
{
  "received": 4,
  "processed": 2
}
```

### `len(s.jobs)`

`jobs` bir channel'dır. `len` channel içinde o anda bekleyen değer sayısını
verir. Bu, çalışan worker'ların işlediği işi değil, kuyruğa alınmış ama henüz
worker tarafından alınmamış işleri gösterir.

---

## 7. Struct ve JSON etiketleri

```go
type Job struct {
    ID        int       `json:"id"`
    Payload   string    `json:"payload"`
    CreatedAt time.Time `json:"created_at"`
}
```

`struct`, ilişkili alanları tek bir tür altında toplar. `Job` artık kendi
türümüzdür:

```go
job := Job{
    ID:      1,
    Payload: "email gönder",
}
```

Backtick arasındaki `json:"id"` bir **struct tag**'dir. `encoding/json`
paketine bu alanın JSON'da `id` ismiyle görünmesini söyler.

---

## 8. Pointer (`*`) nedir?

```go
number := 10
pointer := &number
fmt.Println(*pointer) // 10
```

- `&number`: `number` değişkeninin adresini alır.
- `*pointer`: adresteki değere erişir.

Pointer kullanınca büyük struct'ı kopyalamak yerine aynı nesneye referans
verebiliriz. `*Server`, `*http.Request`, `*sync.WaitGroup` örnekleri bu
projede bunu yapar.

Go'da pointer arithmetic yoktur; pointer'ı bir sonraki belleğe elle ilerletme
gibi işlemler yapılamaz.

---

## 9. Interface nedir?

Interface, bir türün hangi method'lara sahip olması gerektiğini tanımlar.
`http.Handler` kabaca şöyledir:

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Bir tür bu method'a sahipse `http.Handler` olarak kullanılabilir. Interface'i
uyguladığını ayrıca `implements` diye yazmak gerekmez; Go bunu otomatik anlar.

Bu yüzden kendi middleware'imiz:

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        next.ServeHTTP(w, r)
    })
}
```

`http.HandlerFunc`, fonksiyon türünü `http.Handler` interface'ine çeviren
standart bir adaptördür.

---

## 10. Channel ve `select`

Channel, goroutine'ler arasında veri taşıyan typed bir iletişim aracıdır:

```go
jobs := make(chan Job, 10)
jobs <- job       // channel'a gönder
job := <-jobs     // channel'dan al
```

`make(chan Job, 10)` kapasitesi 10 olan buffered channel oluşturur.

- Kapasite doluysa gönderme bekler.
- Channel boşsa alma bekler.
- Bu davranış doğal bir backpressure oluşturur.

Worker'daki yapı:

```go
select {
case <-ctx.Done():
    return
case job := <-s.jobs:
    process(job)
}
```

`select`, birden fazla channel operasyonundan hangisi önce hazır olursa onu
çalıştırır. Burada worker ya iptal sinyalini alır ya da iş kuyruğundan job alır.

### Channel yönleri

Fonksiyon imzasında channel yönünü sınırlayabiliriz:

```go
func sendOnly(out chan<- Job) {}
func receiveOnly(in <-chan Job) {}
```

- `chan<- Job`: sadece gönderme
- `<-chan Job`: sadece alma
- `chan Job`: gönderme ve alma

---

## 11. `context` ne işe yarar?

`context.Context`, bir işin ne zaman iptal edileceğini veya deadline'ını
goroutine'lere taşır.

```go
ctx, stop := signal.NotifyContext(
    context.Background(),
    os.Interrupt,
    syscall.SIGTERM,
)
defer stop()
```

Docker container'ı durdurulurken genellikle `SIGTERM` gönderilir. Bu sinyal
context'i iptal eder. Worker'lar:

```go
case <-ctx.Done():
    return
```

ile bunu görür ve çıkar. `Done()` kapanan bir channel döndürür.

`context.Background()` başlangıç/root context'tir. Alt operasyonlarda
`context.WithTimeout` veya `context.WithCancel` kullanılır.

---

## 12. Projede kullanılan standart kütüphaneler

Go'nun standart kütüphanesi dış paket kurmadan gelen paketlerdir.
`go.mod` içinde bağımlılık olmamasının sebebi budur.

### `context`

İptal, timeout ve request yaşam döngüsünü taşır. Network çağrılarında
unutulmuş/sonsuz beklemeleri önlemeye yardım eder.

### `encoding/json`

Go struct/map değerlerini JSON'a çevirir ve JSON'u Go değerlerine parse eder.

```go
json.NewDecoder(r.Body).Decode(&input)
json.NewEncoder(w).Encode(value)
```

### `errors`

Hataları karşılaştırmak ve sarmalanmış hataları kontrol etmek için kullanılır.

```go
errors.Is(err, http.ErrServerClosed)
```

### `log`

Zaman damgalı log yazmak için kullanılır:

```go
log.Printf("worker=%d started", id)
```

### `net/http`

HTTP server, request, response, route ve middleware altyapısını sağlar.

- `http.NewServeMux()`: route yönlendirici
- `mux.HandleFunc(...)`: handler kaydı
- `http.Server`: server ayarları
- `http.StatusOK`: 200 sabiti
- `http.ListenAndServe()`: server'ı dinlemeye başlar
- `http.Server.Shutdown()`: kontrollü kapanış

### `os`

İşletim sistemiyle iletişim kurar. Bu projede environment variable okur:

```go
os.Getenv("PORT")
```

### `os/signal`

İşletim sistemi sinyallerini yakalamaya yarar. `NotifyContext`, sinyal
geldiğinde context'i iptal eder.

### `strconv`

String ile sayısal türler arasında dönüşüm yapar:

```go
strconv.Atoi("8080")
strconv.Itoa(8080)
```

### `sync`

Goroutine'ler arasında senkronizasyon sağlar:

- `sync.WaitGroup`: goroutine'lerin bitmesini bekler.
- `sync.RWMutex`: okuma/yazma erişimini kilitler.

### `syscall`

İşletim sistemi sinyalleri gibi düşük seviyeli işletim sistemi sabitlerine
erişir. Bu projede `syscall.SIGTERM` kullanılır.

### `time`

Zaman, süre ve timeout işlemleri içindir:

```go
time.Now().UTC()
time.After(2 * time.Second)
context.WithTimeout(..., 10*time.Second)
```

---

## 13. Go'da kontrol yapıları

### `if`

Parantez gerekmez:

```go
if err != nil {
    return
}
```

Başlangıç ifadesi ile:

```go
if err := doWork(); err != nil {
    return err
}
```

### `for`

Go'da tek döngü anahtar kelimesi `for`dur:

```go
for id := 1; id <= workerCount; id++ {
    // ...
}
```

Sonsuz döngü:

```go
for {
    // select ile çıkış koşulu beklenebilir.
}
```

### `switch`

```go
switch status {
case "ok":
    log.Println("başarılı")
default:
    log.Println("bilinmeyen")
}
```

---

## 14. Kapsam (scope) ve görünürlük

Go'da küçük harfle başlayan isimler paket dışından görünmez:

```go
func writeJSON() {} // sadece package main içinde
func NewServer() {} // başka paketlerden de kullanılabilir
```

Büyük harfle başlamak exported/public anlamına gelir. Değişkenlerin kapsamı
ise tanımlandıkları blokla sınırlıdır.

```go
if err := doWork(); err != nil {
    // err burada vardır
}
// err burada yoktur
```

---

## 15. Bu projeyi çalışırken nasıl incelemeli?

Önce şu komutları çalıştırın:

```bash
go run .
curl http://localhost:8080/stats
curl -X POST http://localhost:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{"payload":"birinci iş"}'
```

Sonra farklı terminallerden aynı anda birkaç `POST` isteği gönderin. Loglarda
farklı worker'ların işleri aldığını gözlemleyin.

Race detector ile test edin:

```bash
go test -race ./...
```

Bu proje için en önemli deneyler:

1. `WORKERS=1` ile `WORKERS=5` arasındaki farkı gözlemleyin.
2. `QUEUE_SIZE=1` verip çok sayıda iş gönderin.
3. İşlenme sırasında Docker container'ı durdurun ve shutdown loglarını okuyun.
4. `stats` içindeki kilidi geçici olarak kaldırıp `go test -race` davranışını
   ileride yazacağınız testlerle inceleyin.

