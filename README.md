# ⚡ Binance & DEX Terminal (`binance-terminal`)

Terminal TUI (*Terminal User Interface*) interaktif, cepat, dan modern untuk memantau harga cryptocurrency secara *real-time*, baik **koin major**, **stablecoin**, maupun **meme coin DEX (Solana, Base, Ethereum, BSC)**.

Dirancang dengan prinsip **SLC (Simple, Lovable, Complete)**:
- **Simple**: Langsung jalan dalam 1 detik tanpa konfigurasi awal. Tidak wajib API key untuk monitoring harga publik.
- **Lovable**: Tampilan modern dengan palet warna neon, sparkline grafik tren harga, pembaruan real-time, dan animasi indikator harga naik/turun.
- **Complete**: Streaming WebSocket Binance, integrasi DexScreener untuk meme coin baru, integrasi Signed API Binance untuk melacak saldo & portofolio spot, alert harga, dan penyimpanan konfigurasi otomatis.

---

```
⚡ CRYPTO TERMINAL     [1] Watchlist     [2] Meme Radar     [3] Portfolio     [4] Detail    ● BNC WS  ● DEX  [USD]  07:55:00 UTC
──────────────────────────────────────────────────────────────────────────────────────────────────────────────────
TICKER        SRC    CAT    PRICE          24H CHG     24H HIGH     24H LOW      VOLUME      TREND
▶ BTC          BNC   MAJOR  $77,820.00     + 1.75% ▲   $77,850.00   $76,000.00   $1.04B      ⡠⠤⠒⠊⠉
  ETH          BNC   MAJOR  $2,493.40      + 2.02% ▲   $2,498.68    $2,428.12    $654.80M    ⡠⠤⠒⠊⠉
  SOL          BNC   MAJOR  $106.06        + 6.00% ▲   $106.12      $99.53       $302.75M    ⣀⠤⠒⠊⠉
  BNB          BNC   MAJOR  $753.81        + 3.97% ▲   $759.31      $720.89      $126.86M    ⡠⠤⠒⠊⠉
  USDC         BNC   STABL  $1.00           -0.01% ▼   $1.00        $1.00        $2.80B      ━━━━━━━━━━
  FDUSD        BNC   STABL  $0.9991        + 0.04% ▲   $0.9994      $0.9984      $29.30M     ━━━━━━━━━━
  DOGE         BNC   MEME   $0.0844        + 3.99% ▲   $0.0847      $0.0806      $53.48M     ⡠⠤⠒⠊⠉
  PEPE         BNC   MEME   $0.00000368    + 5.44% ▲   $0.00000370  $0.00000345  $25.53M     ⡠⠤⠒⠊⠉
  WIF          BNC   MEME   $2.14           + 5.30% ▲   $2.25        $1.98        $340.00M    ⡠⠤⠒⠊⠉
  POPCAT (SOL) DEX   MEME   $0.8520        - 3.20% ▼   $0.9100      $0.8300      $95.00M     ⠉⠉⠒⠤⣀
──────────────────────────────────────────────────────────────────────────────────────────────────────────────────
[a] Add  [d] Remove  [/] Filter  [s] Sort  [c] Config  [!] Alert  [r] Refresh  [?] Help  [q] Quit
```

---

## 🚀 Cara Instalasi untuk Semua Orang

Aplikasi ini dapat diinstal dan dijalankan dengan berbagai cara sesuai kenyamanan Anda:

### Metode 1: Instal Cepat via Makefile / Script (Rekomendasi)
```bash
# Clone repository
git clone https://github.com/binance-terminal/binance-terminal.git
cd binance-terminal

# Jalankan installer otomatis (meng-compile dan menaruh binary ke /usr/local/bin)
./scripts/install.sh

# Atau menggunakan make
make install
```
Setelah itu, Anda bisa langsung mengetik `binance-terminal` di terminal mana saja.

### Metode 2: Jalankan Instan via NPX (Pengguna Node.js)
```bash
npx binance-terminal
# Atau instal secara global
npm install -g binance-terminal
```

### Metode 3: Build Mandiri dengan Go
```bash
go build -ldflags="-s -w" -o binance-terminal cmd/binance-terminal/main.go
./binance-terminal
```

---

## 🎮 Navigasi & Shortcut Keyboard

| Tombol | Fungsi |
| :--- | :--- |
| `Tab` / `1 - 4` | Pindah tab: `[1] Watchlist`, `[2] Meme Radar`, `[3] Portfolio`, `[4] Detail View` |
| `↑` / `↓` atau `j` / `k` | Navigasi kursor naik/turun memilih koin |
| `Enter` | Buka tampilan detail statistik koin yang sedang dipilih |
| `a` | **Tambah koin baru** ke Watchlist (otomatis deteksi simbol Binance atau Contract Address DEX) |
| `d` atau `x` | **Hapus koin** yang dipilih dari Watchlist |
| `/` | **Filter / Cari cepat** koin di watchlist yang sedang aktif |
| `s` | **Urutkan (Sort)** daftar berdasarkan: Default, Kenaikan 24h (%), Harga, Volume, atau Nama |
| `c` | **Buka menu Konfigurasi**: Masukkan API Key Binance, ganti mata uang (USD / IDR), toggle suara alert |
| `!` | **Pasang Price Alert** pada koin yang dipilih (bunyi bell terminal saat harga tersentuh) |
| `r` | **Refresh paksa** semua data harga dan saldo portofolio |
| `?` atau `h` | Buka layar bantuan shortcut lengkap |
| `q` / `Ctrl + C` | Keluar dari aplikasi |

---

## 🔑 Konfigurasi API Binance (Opsional)

> **Catatan Penting**: Anda **TIDAK PERLU** memasukkan API key jika hanya ingin memantau harga publik Binance dan koin meme DEX. Data harga publik langsung berjalan otomatis.

API Key hanya diperlukan jika Anda ingin menggunakan **Tab [3] Portfolio** untuk memantau saldo spot Binance Anda.

### Cara Memasukkan API Key:
1. Jalankan `binance-terminal`.
2. Tekan tombol **`c`** pada keyboard untuk membuka menu konfigurasi.
3. Masukkan **Binance API Key** dan **Binance API Secret**.
4. *(Tips Keamanan)*: Di akun Binance Anda, buat API Key dengan izin **hanya Baca (Read-Only)**. Jangan aktifkan izin Spot Trading atau Withdrawal.
5. Tekan `Tab` lalu pilih **[ SAVE & APPLY ]** dan tekan `Enter`.
6. Tekan `3` untuk membuka Tab Portfolio dan melihat saldo aset crypto Anda secara real-time.

Kredensial disimpan secara aman di mesin lokal Anda pada:
```
~/.config/binance-terminal/config.json
```

---

## 💎 Memantau Koin Meme DEX (Solana, Base, Ethereum)

Koin meme yang belum listing di Binance tetap bisa dipantau langsung!

1. Tekan tombol **`a`** untuk membuka menu tambah koin.
2. Masukkan simbol koin atau **paste Smart Contract Address (CA)** token tersebut (contoh: mint address token Solana dari Pump.fun atau Raydium).
3. Pilih Data Source: `DEX (DexScreener)` atau `Auto-Detect`.
4. Tekan `Enter`. Aplikasi akan otomatis mengambil data harga USD, likuiditas, FDV / Market Cap, dan pergerakan 24 jam langsung dari DexScreener.

---

## 🇮🇩 Dukungan Anti-Blokir / Bebas VPN (Indonesia & Global)

Secara default, aplikasi menggunakan endpoint resmi **Binance Vision Mirror** (`https://data-api.binance.vision` dan `wss://data-stream.binance.vision:9443`) untuk data publik. 

Artinya, aplikasi ini dapat langsung digunakan oleh pengguna di Indonesia (Telkomsel, Indihome, Biznet, Starlink, dll.) **tanpa perlu menggunakan VPN**!

---

## 🛠️ Struktur Proyek

```
binance-terminal/
├── cmd/
│   └── binance-terminal/
│       └── main.go              # Entry point aplikasi
├── internal/
│   ├── api/
│   │   ├── binance_ws.go        # Binance WebSocket Client (real-time stream)
│   │   ├── binance_rest.go      # Binance REST Client & Signed Account Portfolio
│   │   └── dexscreener.go       # DexScreener Client (Meme coin multi-chain)
│   ├── config/
│   │   └── config.go            # Penyimpanan ~/.config/binance-terminal/config.json
│   ├── model/
│   │   └── types.go             # Struct tipe data pasar, koin, portofolio, dan alert
│   └── ui/
│       ├── model.go             # Bubbletea Model & state event loop
│       ├── view.go              # Lipgloss rendering untuk semua tab
│       ├── styles.go            # Tema warna, formatting harga, & badge
│       ├── sparkline.go         # Generator Unicode/Braille sparkline
│       ├── modal_add.go         # Modal tambah koin
│       ├── modal_config.go      # Modal konfigurasi API & preferensi
│       ├── modal_alert.go       # Modal pasang target alert harga
│       └── view_portfolio.go    # Tampilan portofolio saldo spot
├── scripts/
│   └── install.sh               # Shell script installer
├── bin/
│   └── cli.js                   # Node CLI wrapper untuk npx
├── Makefile                     # Build, install, run, cross-compile
├── package.json                 # Konfigurasi NPM
├── go.mod
└── go.sum
```

---

## 📜 Lisensi
MIT License © 2026.
