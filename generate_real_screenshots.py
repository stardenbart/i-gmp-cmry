import os
import time
from playwright.sync_api import sync_playwright

output_dir = "/Users/apple/System-Audit-/screenshots"
os.makedirs(output_dir, exist_ok=True)

def login_and_capture():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)

        # ----------------------------------------------------
        # 1. Capture Login Page
        # ----------------------------------------------------
        context = browser.new_context(viewport={"width": 1440, "height": 900})
        page = context.new_page()
        page.goto("http://localhost:3000/auth/login")
        page.wait_for_timeout(1500)
        page.screenshot(path=os.path.join(output_dir, "01_login_page.png"))
        print("Captured 01_login_page.png")
        context.close()

        # ----------------------------------------------------
        # 2. Login as ADMIN (admin / admin123)
        # ----------------------------------------------------
        admin_context = browser.new_context(viewport={"width": 1440, "height": 900})
        admin_page = admin_context.new_page()
        admin_page.goto("http://localhost:3000/auth/login")
        admin_page.wait_for_timeout(1000)
        admin_page.fill('input[name="username"]', 'admin')
        admin_page.fill('input[name="password"]', 'admin123')
        admin_page.click('button[type="submit"]')
        admin_page.wait_for_timeout(3000)
        
        print("Admin logged in. Current URL:", admin_page.url)
        
        # Capture Admin Dashboard
        admin_page.screenshot(path=os.path.join(output_dir, "02_admin_dashboard.png"))
        print("Captured 02_admin_dashboard.png")

        # Get base URL from admin page
        base_admin_url = admin_page.url.rstrip('/')

        # Pages to capture for Admin
        admin_subpages = [
            ("03_inspections_page.png", f"{base_admin_url}/inspections"),
            ("04_issues_page.png", f"{base_admin_url}/issues"),
            ("05_master_data_page.png", f"{base_admin_url}/master"),
            ("06_users_page.png", f"{base_admin_url}/users"),
            ("07_gmp_data_page.png", f"{base_admin_url}/gmp-data"),
            ("08_logs_page.png", f"{base_admin_url}/logs"),
            ("09_settings_page.png", f"{base_admin_url}/settings"),
        ]

        for filename, url in admin_subpages:
            admin_page.goto(url)
            admin_page.wait_for_timeout(2500)
            admin_page.screenshot(path=os.path.join(output_dir, filename))
            print(f"Captured {filename}")

        admin_context.close()

        # ----------------------------------------------------
        # 3. Login as AUDITOR (auditor / auditor123)
        # ----------------------------------------------------
        auditor_context = browser.new_context(viewport={"width": 1440, "height": 900})
        auditor_page = auditor_context.new_page()
        auditor_page.goto("http://localhost:3000/auth/login")
        auditor_page.wait_for_timeout(1000)
        auditor_page.fill('input[name="username"]', 'auditor')
        auditor_page.fill('input[name="password"]', 'auditor123')
        auditor_page.click('button[type="submit"]')
        auditor_page.wait_for_timeout(3000)
        
        print("Auditor logged in. Current URL:", auditor_page.url)
        auditor_page.screenshot(path=os.path.join(output_dir, "10_auditor_dashboard.png"))
        print("Captured 10_auditor_dashboard.png")
        auditor_context.close()

        # ----------------------------------------------------
        # 4. Login as AUDITEE / PIC (amcu / amcuan123)
        # ----------------------------------------------------
        auditee_context = browser.new_context(viewport={"width": 1440, "height": 900})
        auditee_page = auditee_context.new_page()
        auditee_page.goto("http://localhost:3000/auth/login")
        auditee_page.wait_for_timeout(1000)
        auditee_page.fill('input[name="username"]', 'amcu')
        auditee_page.fill('input[name="password"]', 'amcuan123')
        auditee_page.click('button[type="submit"]')
        auditee_page.wait_for_timeout(3000)
        
        print("Auditee logged in. Current URL:", auditee_page.url)
        auditee_page.screenshot(path=os.path.join(output_dir, "11_auditee_dashboard.png"))
        print("Captured 11_auditee_dashboard.png")
        auditee_context.close()

        browser.close()
        print("REAL SCREENSHOTS SUITE CAPTURED SUCCESSFULLY!")

if __name__ == "__main__":
    login_and_capture()
