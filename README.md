# mkvs

Go 製のキーバリューストア実装と、YCSB ライクなベンチマークツールです。

## 構成

```text
kvs/        KVS インターフェースと実装（HashMap など）
benchmark/  ベンチマークランナー・統計処理
cmd/bench/  ベンチマーク CLI
reports/    ベンチマーク結果レポート
```

## ベンチマークの実行

### 基本

```bash
go run ./cmd/bench
```

### オプション

| フラグ | デフォルト | 説明 |
| --- | --- | --- |
| `-workload` | `A` | ワークロード（A/B/C/D/F） |
| `-all` | `false` | 全ワークロードを順番に実行 |
| `-store` | `all` | KVS 実装（`all` / `mutex` / `rwmutex`） |
| `-workers` | `0` | 並列ワーカー数（`0` = 1/2/4/8 のスケーリング計測） |
| `-records` | `100000` | ロードフェーズで投入するレコード数 |
| `-ops` | `500000` | 実行するオペレーション数 |
| `-valuesize` | `100` | バリューのサイズ（バイト） |
| `-dist` | `zipfian` | キー分布（`zipfian` / `uniform`） |

### 実行例

```bash
# ワークロード A を 4 ワーカーで実行
go run ./cmd/bench -workload A -workers 4

# 全ワークロードを実行
go run ./cmd/bench -all

# uniform 分布で実行
go run ./cmd/bench -all -dist uniform
```

## ワークロード

| ワークロード | 内容 |
| --- | --- |
| A | 50% Read / 50% Update（update-heavy） |
| B | 95% Read / 5% Update（read-mostly） |
| C | 100% Read（read-only） |
| D | 95% Read / 5% Insert（read-latest） |
| F | 50% Read / 50% RMW（read-modify-write） |

## レポート

実行ごとに `reports/` にテキストファイルが生成されます。ファイル名にはタイムスタンプと git の参照（タグまたはコミット ID）が含まれます。

```text
reports/bench_20260524_143021_abc1234.txt
```

未コミットの変更がある状態で実行した場合は `-dirty` が付きます（`.gitignore` で除外済み）。

```text
reports/bench_20260524_143021_abc1234-dirty.txt  ← git 管理外
```

レポート内には KVS 実装名（`HashMap/Mutex`、`HashMap/RWMutex` など）が記録されるため、複数実装の比較結果を1ファイルにまとめられます。

### bench.sh

`bench.sh` はフラグをそのまま渡せるショートカットです。

```bash
./bench.sh -all
./bench.sh -store rwmutex -workers 4
```

重要な結果を残したいときは手動でコミットします。

```bash
git add reports/
git commit -m "bench: ..."
```
