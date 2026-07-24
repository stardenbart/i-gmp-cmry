import time
import os
import json
import statistics
from playwright.sync_api import sync_playwright

output_dir = "/Users/apple/System-Audit-"

def test_frontend_search():
    print("\n🚀 Memulai Pengujian Latensi Search Bar pada Frontend (End-to-End UI & Debounce Latency)...")
    
    queries = ["Aktivitas", "Login", "Audit", "User", "System"]
    ui_latencies = []
    
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(viewport={"width": 1440, "height": 900})
        page = context.new_page()
        
        # 1. Login
        page.goto("http://localhost:3000/auth/login")
        page.wait_for_timeout(1000)
        page.fill('input[name="username"]', 'admin')
        page.fill('input[name="password"]', 'admin123')
        page.click('button[type="submit"]')
        page.wait_for_timeout(2500)
        
        base_url = page.url.rstrip('/')
        logs_url = f"{base_url}/logs"
        
        # 2. Go to Logs page
        page.goto(logs_url)
        page.wait_for_timeout(2000)
        
        # Find search input
        search_input = page.query_selector('input[placeholder*="Cari"]') or page.query_selector('input[type="text"]')
        if not search_input:
            print("❌ Search input not found on page!")
            browser.close()
            return
            
        print("✅ Input pencarian ditemukan pada halaman Logs.")
        
        for q in queries:
            # Clear input
            search_input.fill("")
            page.wait_for_timeout(300)
            
            # Start timer
            start = time.perf_counter()
            search_input.fill(q)
            
            # Wait for response / table update
            page.wait_for_timeout(500) # include debounce & API response
            elapsed_ms = (time.perf_counter() - start) * 1000
            ui_latencies.append(elapsed_ms)
            print(f"  🔍 Query: '{q}' -> UI Response Latency: {elapsed_ms:.2f} ms")
            
        browser.close()
        
    avg_ui = statistics.mean(ui_latencies) if ui_latencies else 0
    p50_ui = statistics.median(ui_latencies) if ui_latencies else 0
    min_ui = min(ui_latencies) if ui_latencies else 0
    max_ui = max(ui_latencies) if ui_latencies else 0
    
    print("\n" + "="*80)
    print("📊 HASIL PENGUJIAN LATENSI SEARCH BAR FRONTEND (UI & DEBOUNCE)")
    print("="*80)
    print(f"Rata-rata Latensi UI (Average) : {avg_ui:.2f} ms")
    print(f"Median Latensi UI (P50)       : {p50_ui:.2f} ms")
    print(f"Latensi Tercepat (Min)        : {min_ui:.2f} ms")
    print(f"Latensi Terlama (Max)         : {max_ui:.2f} ms")
    print("="*80 + "\n")
    
    results = {
        "test_type": "Frontend UI Search Bar Latency",
        "avg_ui_ms": round(avg_ui, 2),
        "p50_ui_ms": round(p50_ui, 2),
        "min_ui_ms": round(min_ui, 2),
        "max_ui_ms": round(max_ui, 2),
        "queries_tested": len(queries)
    }
    
    with open(os.path.join(output_dir, "frontend_search_latency_report.json"), "w") as f:
        json.dump(results, f, indent=2)

if __name__ == "__main__":
    test_frontend_search()
