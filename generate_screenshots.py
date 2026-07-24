import os
import json
from playwright.sync_api import sync_playwright

output_dir = "/Users/apple/System-Audit-/screenshots"
os.makedirs(output_dir, exist_ok=True)

with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)

    # 1. Login Page
    page = browser.new_page(viewport={"width": 1440, "height": 900})
    page.goto("http://localhost:3000/auth/login")
    page.wait_for_timeout(1500)
    page.screenshot(path=os.path.join(output_dir, "01_login_page.png"))
    print("Captured 01_login_page.png")
    page.close()

    # 2. Admin Role Pages
    admin_state = {
        "state": {
            "user": {
                "id": "USR-ADMIN-01",
                "name": "Bimo Bintang",
                "email": "admin@cimory.com",
                "role_id": "RL-ADMIN",
                "role_name": "Administrator"
            },
            "token": "valid-admin-token"
        },
        "version": 0
    }
    
    init_script = """
    localStorage.setItem('auth-storage', JSON.stringify(""" + json.dumps(admin_state) + """));
    document.cookie = 'auth-token=valid-token; path=/';
    document.cookie = 'user-id=USR-ADMIN-01; path=/';
    """
    
    context = browser.new_context(viewport={"width": 1440, "height": 900})
    context.add_init_script(init_script)
    
    admin_page = context.new_page()
    
    pages_to_capture = [
        ("02_admin_dashboard.png", "http://localhost:3000/cimory/dashboard/USR-ADMIN-01"),
        ("03_inspections_page.png", "http://localhost:3000/cimory/dashboard/USR-ADMIN-01/inspections"),
        ("04_issues_page.png", "http://localhost:3000/cimory/dashboard/USR-ADMIN-01/issues"),
        ("05_master_data_page.png", "http://localhost:3000/cimory/dashboard/USR-ADMIN-01/master"),
        ("06_users_page.png", "http://localhost:3000/cimory/dashboard/USR-ADMIN-01/users"),
        ("07_gmp_data_page.png", "http://localhost:3000/cimory/dashboard/USR-ADMIN-01/gmp-data"),
        ("08_logs_page.png", "http://localhost:3000/cimory/dashboard/USR-ADMIN-01/logs"),
        ("09_settings_page.png", "http://localhost:3000/cimory/dashboard/USR-ADMIN-01/settings"),
    ]

    for filename, url in pages_to_capture:
        admin_page.goto(url)
        admin_page.wait_for_timeout(2000)
        admin_page.screenshot(path=os.path.join(output_dir, filename))
        print(f"Captured {filename}")
        
    context.close()

    # 3. Auditor Role Page
    auditor_state = {
        "state": {
            "user": {
                "id": "USR-AUDITOR-01",
                "name": "Dewi Sartika (Auditor QA)",
                "email": "auditor@cimory.com",
                "role_id": "RL-AUDITOR",
                "role_name": "Auditor"
            },
            "token": "valid-auditor-token"
        },
        "version": 0
    }
    auditor_script = """
    localStorage.setItem('auth-storage', JSON.stringify(""" + json.dumps(auditor_state) + """));
    document.cookie = 'auth-token=valid-token; path=/';
    document.cookie = 'user-id=USR-AUDITOR-01; path=/';
    """
    auditor_context = browser.new_context(viewport={"width": 1440, "height": 900})
    auditor_context.add_init_script(auditor_script)
    auditor_page = auditor_context.new_page()
    auditor_page.goto("http://localhost:3000/cimory/dashboard/USR-AUDITOR-01")
    auditor_page.wait_for_timeout(2000)
    auditor_page.screenshot(path=os.path.join(output_dir, "10_auditor_dashboard.png"))
    print("Captured 10_auditor_dashboard.png")
    auditor_context.close()

    # 4. Auditee / PIC Role Page
    auditee_state = {
        "state": {
            "user": {
                "id": "USR-PIC-01",
                "name": "Rahmat Hidayat (PIC Area)",
                "email": "pic@cimory.com",
                "role_id": "RL-AUDITEE",
                "role_name": "Auditee / PIC"
            },
            "token": "valid-auditee-token"
        },
        "version": 0
    }
    auditee_script = """
    localStorage.setItem('auth-storage', JSON.stringify(""" + json.dumps(auditee_state) + """));
    document.cookie = 'auth-token=valid-token; path=/';
    document.cookie = 'user-id=USR-PIC-01; path=/';
    """
    auditee_context = browser.new_context(viewport={"width": 1440, "height": 900})
    auditee_context.add_init_script(auditee_script)
    auditee_page = auditee_context.new_page()
    auditee_page.goto("http://localhost:3000/cimory/dashboard/USR-PIC-01")
    auditee_page.wait_for_timeout(2000)
    auditee_page.screenshot(path=os.path.join(output_dir, "11_auditee_dashboard.png"))
    print("Captured 11_auditee_dashboard.png")
    auditee_context.close()

    browser.close()
    print("All screenshots generated successfully!")
