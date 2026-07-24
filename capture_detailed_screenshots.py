import os
import time
import json
from playwright.sync_api import sync_playwright

output_dir = "/Users/apple/System-Audit-/screenshots"
os.makedirs(output_dir, exist_ok=True)

def capture_detailed():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)

        # ----------------------------------------------------
        # 1. Login Page
        # ----------------------------------------------------
        context = browser.new_context(viewport={"width": 1440, "height": 900})
        page = context.new_page()
        page.goto("http://localhost:3000/auth/login")
        page.wait_for_timeout(1500)
        page.screenshot(path=os.path.join(output_dir, "01_login_page.png"))
        print("Captured 01_login_page.png")
        context.close()

        # ----------------------------------------------------
        # 2. ADMIN ROLE (admin / admin123)
        # ----------------------------------------------------
        admin_ctx = browser.new_context(viewport={"width": 1440, "height": 900})
        admin_page = admin_ctx.new_page()
        admin_page.goto("http://localhost:3000/auth/login")
        admin_page.wait_for_timeout(1000)
        admin_page.fill('input[name="username"]', 'admin')
        admin_page.fill('input[name="password"]', 'admin123')
        admin_page.click('button[type="submit"]')
        admin_page.wait_for_timeout(3000)

        base_admin_url = admin_page.url.rstrip('/')
        print("Admin base URL:", base_admin_url)

        # 2a. Admin Dashboard - Top Section
        admin_page.evaluate("window.scrollTo(0, 0)")
        admin_page.wait_for_timeout(1000)
        admin_page.screenshot(path=os.path.join(output_dir, "02_admin_dashboard_top.png"))
        print("Captured 02_admin_dashboard_top.png")

        # 2b. Admin Dashboard - Monitoring per Role Section
        admin_page.evaluate("window.scrollTo(0, 450)")
        admin_page.wait_for_timeout(1000)
        admin_page.screenshot(path=os.path.join(output_dir, "03_admin_dashboard_monitoring.png"))
        print("Captured 03_admin_dashboard_monitoring.png")

        # 2c. Admin Dashboard - Chart Section
        admin_page.evaluate("window.scrollTo(0, 950)")
        admin_page.wait_for_timeout(1000)
        admin_page.screenshot(path=os.path.join(output_dir, "04_admin_dashboard_chart.png"))
        print("Captured 04_admin_dashboard_chart.png")

        # 2d. Inspections List Page
        admin_page.goto(f"{base_admin_url}/inspections")
        admin_page.wait_for_timeout(2000)
        admin_page.screenshot(path=os.path.join(output_dir, "05_admin_inspections_list.png"))
        print("Captured 05_admin_inspections_list.png")

        # 2e. Issues List Page
        admin_page.goto(f"{base_admin_url}/issues")
        admin_page.wait_for_timeout(2000)
        admin_page.screenshot(path=os.path.join(output_dir, "06_admin_issues_list.png"))
        print("Captured 06_admin_issues_list.png")

        # 2f. WO / WR Page
        admin_page.goto(f"{base_admin_url}/wowr")
        admin_page.wait_for_timeout(2000)
        admin_page.screenshot(path=os.path.join(output_dir, "07_admin_wowr_page.png"))
        print("Captured 07_admin_wowr_page.png")

        # 2g. Master Data - Area/Kawasan Tab
        admin_page.goto(f"{base_admin_url}/master")
        admin_page.wait_for_timeout(2000)
        admin_page.screenshot(path=os.path.join(output_dir, "08_admin_master_area.png"))
        print("Captured 08_admin_master_area.png")

        # Click Aspek Tab in Master Data if exists
        try:
          aspek_btn = admin_page.query_selector("button:has-text('Aspek')")
          if aspek_btn:
              aspek_btn.click()
              admin_page.wait_for_timeout(1500)
              admin_page.screenshot(path=os.path.join(output_dir, "09_admin_master_aspek.png"))
              print("Captured 09_admin_master_aspek.png")

          uraian_btn = admin_page.query_selector("button:has-text('Uraian')")
          if uraian_btn:
              uraian_btn.click()
              admin_page.wait_for_timeout(1500)
              admin_page.screenshot(path=os.path.join(output_dir, "10_admin_master_uraian.png"))
              print("Captured 10_admin_master_uraian.png")
        except Exception as e:
          print("Master data subtabs note:", e)

        # 2h. Users Management Page
        admin_page.goto(f"{base_admin_url}/users")
        admin_page.wait_for_timeout(2000)
        admin_page.screenshot(path=os.path.join(output_dir, "11_admin_users_list.png"))
        print("Captured 11_admin_users_list.png")

        # 2i. GMP Data Page
        admin_page.goto(f"{base_admin_url}/gmp-data")
        admin_page.wait_for_timeout(2000)
        admin_page.screenshot(path=os.path.join(output_dir, "12_admin_gmp_data.png"))
        print("Captured 12_admin_gmp_data.png")

        # 2j. Logs Page - Activity Tab
        admin_page.goto(f"{base_admin_url}/logs")
        admin_page.wait_for_timeout(2000)
        admin_page.screenshot(path=os.path.join(output_dir, "13_admin_logs_activity.png"))
        print("Captured 13_admin_logs_activity.png")

        # Logs Page - Login Tab
        try:
          login_tab_btn = admin_page.query_selector("button:has-text('Riwayat Sesi Masuk')") or admin_page.query_selector("button:has-text('Login Logs')")
          if login_tab_btn:
              login_tab_btn.click()
              admin_page.wait_for_timeout(1500)
              admin_page.screenshot(path=os.path.join(output_dir, "14_admin_logs_login.png"))
              print("Captured 14_admin_logs_login.png")
        except Exception as e:
          print("Logs subtab note:", e)

        # 2k. Settings Page - General Tab
        admin_page.goto(f"{base_admin_url}/settings")
        admin_page.wait_for_timeout(2000)
        admin_page.screenshot(path=os.path.join(output_dir, "15_admin_settings_general.png"))
        print("Captured 15_admin_settings_general.png")

        # Settings Page - Email Templates Tab
        try:
          email_tab_btn = admin_page.query_selector("button:has-text('Template Email')")
          if email_tab_btn:
              email_tab_btn.click()
              admin_page.wait_for_timeout(1500)
              admin_page.screenshot(path=os.path.join(output_dir, "16_admin_settings_email.png"))
              print("Captured 16_admin_settings_email.png")
        except Exception as e:
          print("Settings subtab note:", e)

        admin_ctx.close()

        # ----------------------------------------------------
        # 3. AUDITOR ROLE (auditor / auditor123)
        # ----------------------------------------------------
        auditor_ctx = browser.new_context(viewport={"width": 1440, "height": 900})
        auditor_page = auditor_ctx.new_page()
        auditor_page.goto("http://localhost:3000/auth/login")
        auditor_page.wait_for_timeout(1000)
        auditor_page.fill('input[name="username"]', 'auditor')
        auditor_page.fill('input[name="password"]', 'auditor123')
        auditor_page.click('button[type="submit"]')
        auditor_page.wait_for_timeout(3000)

        auditor_page.evaluate("window.scrollTo(0, 0)")
        auditor_page.wait_for_timeout(1000)
        auditor_page.screenshot(path=os.path.join(output_dir, "17_auditor_dashboard_overview.png"))
        print("Captured 17_auditor_dashboard_overview.png")

        auditor_page.evaluate("window.scrollTo(0, 450)")
        auditor_page.wait_for_timeout(1000)
        auditor_page.screenshot(path=os.path.join(output_dir, "18_auditor_dashboard_trend.png"))
        print("Captured 18_auditor_dashboard_trend.png")
        auditor_ctx.close()

        # ----------------------------------------------------
        # 4. AUDITEE / PIC ROLE (amcu / amcuan123)
        # ----------------------------------------------------
        auditee_ctx = browser.new_context(viewport={"width": 1440, "height": 900})
        auditee_page = auditee_ctx.new_page()
        auditee_page.goto("http://localhost:3000/auth/login")
        auditee_page.wait_for_timeout(1000)
        auditee_page.fill('input[name="username"]', 'amcu')
        auditee_page.fill('input[name="password"]', 'amcuan123')
        auditee_page.click('button[type="submit"]')
        auditee_page.wait_for_timeout(3000)

        auditee_page.evaluate("window.scrollTo(0, 0)")
        auditee_page.wait_for_timeout(1000)
        auditee_page.screenshot(path=os.path.join(output_dir, "19_auditee_dashboard_overview.png"))
        print("Captured 19_auditee_dashboard_overview.png")

        auditee_page.evaluate("window.scrollTo(0, 450)")
        auditee_page.wait_for_timeout(1000)
        auditee_page.screenshot(path=os.path.join(output_dir, "20_auditee_tasks_list.png"))
        print("Captured 20_auditee_tasks_list.png")
        auditee_ctx.close()

        browser.close()
        print("ALL 20 DETAILED SCREENSHOTS CAPTURED SUCCESSFULLY!")

if __name__ == "__main__":
    capture_detailed()
