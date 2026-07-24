#!/usr/bin/env python3
import time
import statistics
import concurrent.futures
import requests
import json
import os
from datetime import datetime

BASE_URL = os.environ.get("API_BASE_URL", "http://localhost:8080/api/v1")
USERNAME = os.environ.get("API_USERNAME", "admin")
PASSWORD = os.environ.get("API_PASSWORD", "admin123")

SEARCH_QUERIES = [
    "Area", "Inspeksi", "Open", "Draft", "admin", 
    "Aktivitas", "QC", "2026", "Completed", "Cimory",
    "Gagal", "Berhasil", "Master", "User", "Temuan"
]

ENDPOINTS = [
    {"name": "Pencarian Inspeksi (OpenSearch)", "path": "/search/inspections"},
    {"name": "Pencarian Temuan (OpenSearch)", "path": "/search/issues"},
    {"name": "Pencarian Riwayat Log (OpenSearch)", "path": "/search/logs"},
    {"name": "Filter List Inspeksi (GORM/DB)", "path": "/inspections"},
    {"name": "Filter List Temuan (GORM/DB)", "path": "/issues"}
]

def get_auth_token():
    try:
        url = f"{BASE_URL}/auth/login"
        resp = requests.post(url, json={"username": USERNAME, "password": PASSWORD}, timeout=10)
        if resp.status_code == 200:
            token = resp.json().get("data", {}).get("token")
            return token
        print(f"❌ Login failed with status {resp.status_code}: {resp.text}")
        return None
    except Exception as e:
        print(f"❌ Error connecting to server for login: {e}")
        return None

def single_request(endpoint, query, token):
    headers = {"Authorization": f"Bearer {token}"} if token else {}
    url = f"{BASE_URL}{endpoint['path']}"
    params = {"q": query, "limit": 10}
    
    start_time = time.perf_counter()
    try:
        resp = requests.get(url, params=params, headers=headers, timeout=10)
        elapsed_ms = (time.perf_counter() - start_time) * 1000
        return {
            "endpoint": endpoint["name"],
            "path": endpoint["path"],
            "query": query,
            "status_code": resp.status_code,
            "latency_ms": elapsed_ms,
            "success": resp.status_code == 200
        }
    except Exception as e:
        elapsed_ms = (time.perf_counter() - start_time) * 1000
        return {
            "endpoint": endpoint["name"],
            "path": endpoint["path"],
            "query": query,
            "status_code": 0,
            "latency_ms": elapsed_ms,
            "success": False,
            "error": str(e)
        }

def run_latency_benchmark(token, iterations_per_endpoint=20, concurrency=5):
    print(f"\n🚀 Memulai Pengujian Latensi Pencarian ({iterations_per_endpoint * len(ENDPOINTS)} total request, concurrency={concurrency})...\n")
    
    overall_results = {}
    
    for endpoint in ENDPOINTS:
        print(f"⏳ Pengujian: {endpoint['name']} ({endpoint['path']})...")
        tasks = []
        with concurrent.futures.ThreadPoolExecutor(max_workers=concurrency) as executor:
            for i in range(iterations_per_endpoint):
                query = SEARCH_QUERIES[i % len(SEARCH_QUERIES)]
                tasks.append(executor.submit(single_request, endpoint, query, token))
            
            results = [task.result() for task in concurrent.futures.as_completed(tasks)]
        
        latencies = [r["latency_ms"] for r in results]
        successes = [r for r in results if r["success"]]
        
        latencies_sorted = sorted(latencies)
        p50 = statistics.median(latencies_sorted) if latencies_sorted else 0
        p90 = latencies_sorted[int(len(latencies_sorted) * 0.90)] if latencies_sorted else 0
        p95 = latencies_sorted[int(len(latencies_sorted) * 0.95)] if latencies_sorted else 0
        p99 = latencies_sorted[int(len(latencies_sorted) * 0.99)] if latencies_sorted else 0
        
        avg_latency = statistics.mean(latencies) if latencies else 0
        min_latency = min(latencies) if latencies else 0
        max_latency = max(latencies) if latencies else 0
        std_dev = statistics.stdev(latencies) if len(latencies) > 1 else 0
        success_rate = (len(successes) / len(results)) * 100 if results else 0
        
        overall_results[endpoint["name"]] = {
            "path": endpoint["path"],
            "total_requests": len(results),
            "successful_requests": len(successes),
            "success_rate_pct": round(success_rate, 2),
            "avg_ms": round(avg_latency, 2),
            "min_ms": round(min_latency, 2),
            "max_ms": round(max_latency, 2),
            "p50_ms": round(p50, 2),
            "p90_ms": round(p90, 2),
            "p95_ms": round(p95, 2),
            "p99_ms": round(p99, 2),
            "std_dev_ms": round(std_dev, 2)
        }
    
    return overall_results

def print_report(results):
    print("\n" + "="*95)
    print("📊 LAPORAN HASIL PENGUJIAN LATENSI PENCARIAN DATA (SEARCH LATENCY BENCHMARK REPORT)")
    print("="*95)
    print(f"Waktu Pengujian : {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"Target Server   : {BASE_URL}")
    print("-" * 95)
    print(f"{'Endpoint / Fitur':<38} | {'Avg (ms)':<8} | {'P50 (ms)':<8} | {'P90 (ms)':<8} | {'P95 (ms)':<8} | {'Success':<8}")
    print("-" * 95)
    
    for name, stats in results.items():
        print(f"{name:<38} | {stats['avg_ms']:<8.2f} | {stats['p50_ms']:<8.2f} | {stats['p90_ms']:<8.2f} | {stats['p95_ms']:<8.2f} | {stats['success_rate_pct']:<7.1f}%")
    print("="*95 + "\n")

def export_json(results, filepath="search_latency_report.json"):
    with open(filepath, "w") as f:
        json.dump(results, f, indent=2)
    print(f"💾 Laporan hasil pengujian berhasil disimpan ke: {filepath}")

def main():
    print("🔑 Otentikasi ke server backend...")
    token = get_auth_token()
    if not token:
        print("⚠️ Menggunakan pengujian anonim (beberapa endpoint mungkin membutuhkan auth).")
    else:
        print("✅ Otentikasi sukses! Token JWT didapatkan.")
        
    results = run_latency_benchmark(token, iterations_per_endpoint=25, concurrency=5)
    print_report(results)
    export_json(results)

if __name__ == "__main__":
    main()
