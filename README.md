# Go Network & Concurrency Lab

Bu proje, Go'nun network servisleri ve eşzamanlılık özelliklerini kod yazarak öğrenmek için ilk adımdır.

## Çalıştırma

Docker ile:

```bash
docker compose up --build
```

Yerel Go ile:

```bash
go run .
```

Servis `http://localhost:8080` adresinde çalışır.

## API denemeleri

Sağlık kontrolü:

```bash
curl http://localhost:8080/healthz
```

Kuyruğa iş ekleme:

```bash
curl -X POST http://localhost:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{"payload":"ilk Go isim"}'
```

İstatistik:

```bash
curl http://localhost:8080/stats
```

## Bu adımda öğrenilenler

- `net/http`: HTTP server ve route tanımlama
- `go server.worker(...)`: her worker'ı ayrı goroutine olarak çalıştırma
- `chan Job`: handler ile worker'lar arasında güvenli iş kuyruğu
- `select`: channel ve `context` olaylarını birlikte bekleme
- `context`: SIGINT/SIGTERM geldiğinde server ve worker'ları durdurma
- `sync.WaitGroup`: tüm worker'ların bitmesini bekleme
- `sync.RWMutex`: istatistik sayaçlarını eşzamanlı erişime karşı koruma
- Docker multi-stage build: küçük bir runtime image üretme

## Önerilen öğrenme sırası

1. `healthz` handler'ını değiştirip HTTP response üretimini incele.
2. `WORKERS=1` ve `WORKERS=5` ile aynı anda gönderilen işleri gözlemle.
3. `QUEUE_SIZE=1` vererek kuyruğun dolmasını ve handler'ın beklemesini gözlemle.
4. `worker` içindeki `time.After` bölümünü gerçek bir iş fonksiyonuyla değiştir.
5. Sonraki adımda işi `GET /jobs/{id}` ile sorgulanabilir hale getir.
6. Ardından test, Docker healthcheck, PostgreSQL ve mesaj kuyruğu ekle.

Ortam değişkenleri: `PORT` (8080), `WORKERS` (3), `QUEUE_SIZE` (10).
