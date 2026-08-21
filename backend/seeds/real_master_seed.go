package seeds

import (
	"fmt"
	"log"
	"strings"
	"time"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	inspectiondomain "github.com/monitoring-system/backend/internal/domain/inspection"
	issuedomain "github.com/monitoring-system/backend/internal/domain/issue"
	masterdomain "github.com/monitoring-system/backend/internal/domain/master"
	picdomain "github.com/monitoring-system/backend/internal/domain/pic"
	"github.com/monitoring-system/backend/pkg/password"
	"gorm.io/gorm"
)

// SeedRealMasterData populates official master data for Sentul Plant based on user specification.
func SeedRealMasterData(db *gorm.DB) {
	log.Println("🌱 Seeding Real Master Data (Area, Kawasan, DetailKawasan, Aspek, Detail, Uraian, PIC Mapping)...")

	defaultPlantID := "PLT-SENTUL"

	// ── 1. Master Area ────────────────────────────────────────────────────────
	areaDefs := []struct {
		ID   string
		Name string
	}{
		{ID: "AREA-PRODUKSI", Name: "Area Produksi"},
		{ID: "AREA-GUDANG", Name: "Area Gudang"},
		{ID: "AREA-QC", Name: "Area Lab QC & Inline"},
		{ID: "AREA-LUAR", Name: "Area Luar"},
	}

	areaMap := make(map[string]string) // Name -> AreaID
	for _, a := range areaDefs {
		area := masterdomain.Area{
			AreaID:   a.ID,
			PlantID:  &defaultPlantID,
			AreaName: a.Name,
		}
		db.Where("LOWER(\"AreaName\") = LOWER(?) OR \"AreaID\" = ?", a.Name, a.ID).FirstOrCreate(&area)
		areaMap[a.Name] = area.AreaID
	}

	// ── 2. Master Kawasan ─────────────────────────────────────────────────────
	kawasanDefs := []struct {
		Name string
		Area string
	}{
		{Name: "Produksi CMD 1", Area: "Area Produksi"},
		{Name: "Produksi CMD 2", Area: "Area Produksi"},
		{Name: "Produksi CMD 3", Area: "Area Produksi"},
		{Name: "Quality Control", Area: "Area Lab QC & Inline"},
		{Name: "Warehouse RMPM", Area: "Area Gudang"},
		{Name: "Warehouse MS", Area: "Area Gudang"},
		{Name: "Area Luar CMD", Area: "Area Luar"},
		{Name: "Lab R&I", Area: "Area Lab QC & Inline"},
	}

	kawasanMap := make(map[string]string) // Name -> KawasanID
	for i, k := range kawasanDefs {
		areaID := areaMap[k.Area]
		if areaID == "" {
			continue
		}
		kID := fmt.Sprintf("KWS-%03d", i+1)
		kawasan := masterdomain.Kawasan{
			KawasanID:   kID,
			AreaID:      areaID,
			KawasanName: k.Name,
		}
		db.Where("LOWER(\"KawasanName\") = LOWER(?)", k.Name).FirstOrCreate(&kawasan)
		kawasanMap[k.Name] = kawasan.KawasanID
	}

	// ── 3. Detail Kawasan ─────────────────────────────────────────────────────
	detailKawasanDefs := []struct {
		Name    string
		Kawasan string
	}{
		{Name: "Proses", Kawasan: "Produksi CMD 1"},
		{Name: "Filling", Kawasan: "Produksi CMD 1"},
		{Name: "Packing", Kawasan: "Produksi CMD 1"},
		{Name: "Packing", Kawasan: "Produksi CMD 2"},
		{Name: "Proses", Kawasan: "Produksi CMD 2"},
		{Name: "Filling", Kawasan: "Produksi CMD 2"},
		{Name: "Packing", Kawasan: "Produksi CMD 3"},
		{Name: "Proses", Kawasan: "Produksi CMD 3"},
		{Name: "Filling", Kawasan: "Produksi CMD 3"},
		{Name: "Lab Mikrobiologi", Kawasan: "Quality Control"},
		{Name: "Lab Pyschem 1", Kawasan: "Quality Control"},
		{Name: "Lab Pyschem 2 & Inline", Kawasan: "Quality Control"},
		{Name: "Lab Pyschem 3 & Inline", Kawasan: "Quality Control"},
		{Name: "Lab Pyschem 4 & Inline", Kawasan: "Quality Control"},
		{Name: "Bangunan Dalam dan Area Luar CMD", Kawasan: "Area Luar CMD"},
		{Name: "Warehouse RMPM CMD 1", Kawasan: "Warehouse RMPM"},
		{Name: "Warehouse RMPM CMD 2", Kawasan: "Warehouse RMPM"},
		{Name: "Warehouse RMPM CMD 3", Kawasan: "Warehouse RMPM"},
		{Name: "Warehouse Weighing Room", Kawasan: "Warehouse RMPM"},
		{Name: "Warehouse CCIE 3", Kawasan: "Warehouse RMPM"},
		{Name: "Warehouse FG MS Sentul", Kawasan: "Warehouse MS"},
	}

	for i, dk := range detailKawasanDefs {
		kID := kawasanMap[dk.Kawasan]
		if kID == "" {
			continue
		}
		dkID := fmt.Sprintf("DKWS-%03d", i+1)
		detailKawasan := masterdomain.DetailKawasan{
			DetailKawasanID:   dkID,
			KawasanID:         kID,
			DetailKawasanName: dk.Name,
		}
		db.Where("LOWER(\"DetailKawasanName\") = LOWER(?) AND \"KawasanID\" = ?", dk.Name, kID).FirstOrCreate(&detailKawasan)
	}

	// ── 4. Master Aspek ───────────────────────────────────────────────────────
	aspekDefs := []struct {
		Name string
		Area string
	}{
		{Name: "Lingkungan Sarana Produksi - Produksi", Area: "Area Produksi"},
		{Name: "Konstruksi dan Layout Bangunan - Produksi", Area: "Area Produksi"},
		{Name: "Area Produksi - Produksi", Area: "Area Produksi"},
		{Name: "Personil - QC", Area: "Area Lab QC & Inline"},
		{Name: "Fasilitas - QC", Area: "Area Lab QC & Inline"},
		{Name: "Instrumen - QC", Area: "Area Lab QC & Inline"},
		{Name: "Konstruksi dan Infrastruktur Bangunan - QC", Area: "Area Lab QC & Inline"},
		{Name: "Prosedur / Metode - QC", Area: "Area Lab QC & Inline"},
		{Name: "Material - QC", Area: "Area Lab QC & Inline"},
		{Name: "Ketidaksesuaian GMP / 5R Lainnya - QC", Area: "Area Lab QC & Inline"},
		{Name: "Lingkungan Sarana Penyimpanan Material - WH", Area: "Area Gudang"},
		{Name: "Konstruksi dan Layout Bangunan - WH", Area: "Area Gudang"},
		{Name: "Area Penyimpanan Material - WH", Area: "Area Gudang"},
		{Name: "Ketidaksesuaian GMP / 5R Lainnya - WH", Area: "Area Gudang"},
		{Name: "Ketidaksesuaian GMP / 5R Lainnya - Produksi", Area: "Area Produksi"},
		{Name: "Personil - Luar", Area: "Area Luar"},
		{Name: "Lingkungan Sekitar - Luar", Area: "Area Luar"},
		{Name: "Kebersihan Area Luar Gedung - Luar", Area: "Area Luar"},
		{Name: "Konstruksi dan Infrastruktur Bangunan Luar - Luar", Area: "Area Luar"},
		{Name: "Supporting Area Luar - Luar", Area: "Area Luar"},
		{Name: "Ketidaksesuaian GMP / 5R Lainnya - Luar", Area: "Area Luar"},
	}

	aspekMap := make(map[string]string) // Name -> AspekID
	for i, a := range aspekDefs {
		areaID := areaMap[a.Area]
		if areaID == "" {
			continue
		}
		aspID := fmt.Sprintf("ASP-%03d", i+1)
		aspek := masterdomain.Aspek{
			AspekID:   aspID,
			AreaID:    areaID,
			AspekName: a.Name,
		}
		db.Where("LOWER(\"AspekName\") = LOWER(?)", a.Name).FirstOrCreate(&aspek)
		aspekMap[a.Name] = aspek.AspekID
	}

	// ── 5. Detail Pemeriksaan (Detail_Master) ──────────────────────────────────
	detailDefs := []struct {
		Name  string
		Aspek string
	}{
		{Name: "Lokasi Pabrik", Aspek: "Lingkungan Sarana Produksi - Produksi"},
		{Name: "Sarana Jalan", Aspek: "Lingkungan Sarana Produksi - Produksi"},
		{Name: "Lingkungan", Aspek: "Lingkungan Sarana Produksi - Produksi"},
		{Name: "Konstruksi Bangunan", Aspek: "Konstruksi dan Layout Bangunan - Produksi"},
		{Name: "Dinding", Aspek: "Konstruksi dan Layout Bangunan - Produksi"},
		{Name: "Lantai", Aspek: "Konstruksi dan Layout Bangunan - Produksi"},
		{Name: "Atap dan Langit-Langit", Aspek: "Konstruksi dan Layout Bangunan - Produksi"},
		{Name: "Pintu / Rolling Door", Aspek: "Konstruksi dan Layout Bangunan - Produksi"},
		{Name: "Jendela", Aspek: "Konstruksi dan Layout Bangunan - Produksi"},
		{Name: "Perpipaan", Aspek: "Konstruksi dan Layout Bangunan - Produksi"},
		{Name: "Layout Bangunan", Aspek: "Konstruksi dan Layout Bangunan - Produksi"},
		{Name: "Personal Hygiene", Aspek: "Area Produksi - Produksi"},
		{Name: "Supporting", Aspek: "Area Produksi - Produksi"},
		{Name: "Ketidaksesuaian GMP / 5R Lainnya - Produksi", Aspek: "Ketidaksesuaian GMP / 5R Lainnya - Produksi"},
		{Name: "Personal Hygiene", Aspek: "Personil - QC"},
		{Name: "Wastafel", Aspek: "Fasilitas - QC"},
		{Name: "Alat Kebersihan", Aspek: "Fasilitas - QC"},
		{Name: "Lingkungan", Aspek: "Fasilitas - QC"},
		{Name: "Akses Kontrol", Aspek: "Fasilitas - QC"},
		{Name: "Peralatan", Aspek: "Instrumen - QC"},
		{Name: "Chiller", Aspek: "Instrumen - QC"},
		{Name: "Konstruksi", Aspek: "Konstruksi dan Infrastruktur Bangunan - QC"},
		{Name: "Dinding", Aspek: "Konstruksi dan Infrastruktur Bangunan - QC"},
		{Name: "Atap dan Langit-Langit", Aspek: "Konstruksi dan Infrastruktur Bangunan - QC"},
		{Name: "Kaca / Jendela / Pintu", Aspek: "Konstruksi dan Infrastruktur Bangunan - QC"},
		{Name: "Lantai", Aspek: "Konstruksi dan Infrastruktur Bangunan - QC"},
		{Name: "Pencahayaan", Aspek: "Konstruksi dan Infrastruktur Bangunan - QC"},
		{Name: "Kalibrasi", Aspek: "Prosedur / Metode - QC"},
		{Name: "Prosedur Uji", Aspek: "Prosedur / Metode - QC"},
		{Name: "Bahan Baku Material / Bahan Chemical Sanitasi", Aspek: "Material - QC"},
		{Name: "Ketidaksesuaian GMP / 5R Lainnya - QC", Aspek: "Ketidaksesuaian GMP / 5R Lainnya - QC"},
		{Name: "Lokasi Pabrik", Aspek: "Lingkungan Sarana Penyimpanan Material - WH"},
		{Name: "Sarana Jalan", Aspek: "Lingkungan Sarana Penyimpanan Material - WH"},
		{Name: "Lingkungan", Aspek: "Lingkungan Sarana Penyimpanan Material - WH"},
		{Name: "Konstruksi Bangunan", Aspek: "Konstruksi dan Layout Bangunan - WH"},
		{Name: "Dinding", Aspek: "Konstruksi dan Layout Bangunan - WH"},
		{Name: "Lantai", Aspek: "Konstruksi dan Layout Bangunan - WH"},
		{Name: "Atap dan Langit-Langit", Aspek: "Konstruksi dan Layout Bangunan - WH"},
		{Name: "Pintu / Rolling Door", Aspek: "Konstruksi dan Layout Bangunan - WH"},
		{Name: "Jendela", Aspek: "Konstruksi dan Layout Bangunan - WH"},
		{Name: "Perpipaan", Aspek: "Konstruksi dan Layout Bangunan - WH"},
		{Name: "Layout Bangunan", Aspek: "Konstruksi dan Layout Bangunan - WH"},
		{Name: "Personal Hygiene", Aspek: "Area Penyimpanan Material - WH"},
		{Name: "Supporting", Aspek: "Area Penyimpanan Material - WH"},
		{Name: "Ketidaksesuaian GMP / 5R Lainnya - WH", Aspek: "Ketidaksesuaian GMP / 5R Lainnya - WH"},
		{Name: "Personal Hygiene", Aspek: "Personil - Luar"},
		{Name: "Papan Informasi", Aspek: "Lingkungan Sekitar - Luar"},
		{Name: "Sarana Jalan", Aspek: "Lingkungan Sekitar - Luar"},
		{Name: "Drainase", Aspek: "Lingkungan Sekitar - Luar"},
		{Name: "Jalur Pejalan Kaki", Aspek: "Lingkungan Sekitar - Luar"},
		{Name: "Lingkungan", Aspek: "Lingkungan Sekitar - Luar"},
		{Name: "Tempat Sampah", Aspek: "Kebersihan Area Luar Gedung - Luar"},
		{Name: "Konstruksi", Aspek: "Konstruksi dan Infrastruktur Bangunan Luar - Luar"},
		{Name: "Dinding", Aspek: "Konstruksi dan Infrastruktur Bangunan Luar - Luar"},
		{Name: "Atap dan Langit-Langit", Aspek: "Konstruksi dan Infrastruktur Bangunan Luar - Luar"},
		{Name: "Pencahayaan", Aspek: "Konstruksi dan Infrastruktur Bangunan Luar - Luar"},
		{Name: "Perpipaan", Aspek: "Konstruksi dan Infrastruktur Bangunan Luar - Luar"},
		{Name: "Pintu", Aspek: "Konstruksi dan Infrastruktur Bangunan Luar - Luar"},
		{Name: "Jendela", Aspek: "Konstruksi dan Infrastruktur Bangunan Luar - Luar"},
		{Name: "Pest Management", Aspek: "Supporting Area Luar - Luar"},
		{Name: "Penyimpanan Limbah / Reject Produk", Aspek: "Supporting Area Luar - Luar"},
		{Name: "Penyimpanan Sementara Material", Aspek: "Supporting Area Luar - Luar"},
		{Name: "Fasilitas Sanitasi", Aspek: "Supporting Area Luar - Luar"},
		{Name: "Fasilitas dan Peralatan Umum", Aspek: "Supporting Area Luar - Luar"},
		{Name: "Ketidaksesuaian GMP / 5R Lainnya - Luar", Aspek: "Ketidaksesuaian GMP / 5R Lainnya - Luar"},
	}

	detailMap := make(map[string]string) // Key (AspekID + DetailName) -> DetailID
	for i, d := range detailDefs {
		aspID := aspekMap[d.Aspek]
		if aspID == "" {
			continue
		}
		detID := fmt.Sprintf("DET-%03d", i+1)
		detail := masterdomain.Detail{
			DetailID:   detID,
			AspekID:    aspID,
			DetailName: d.Name,
		}
		db.Where("LOWER(\"DetailName\") = LOWER(?) AND \"AspekID\" = ?", d.Name, aspID).FirstOrCreate(&detail)
		detailMap[aspID+"_"+d.Name] = detail.DetailID
	}

	// ── 6. Mapping PIC & User Creation ─────────────────────────────────────────
	picDefs := []struct {
		Area    string
		Kawasan string
		PICName string
	}{
		{Area: "Area Produksi", Kawasan: "Produksi CMD 1", PICName: "Hermaslin Pasaribu"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 1", PICName: "Moh. Rama Gumilar"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 1", PICName: "H.M. Deden Moenawar"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 1", PICName: "Hasan Basri"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 1", PICName: "Muhamad Riski Hardianto ST"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 1", PICName: "Ariansyah"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 1", PICName: "Dzikri Nurazizi"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 1", PICName: "Abdullah Farauk"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 1", PICName: "Nadya Audyra Rahardy"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 2", PICName: "Beben Krisbiantoro"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 2", PICName: "Farhan Lazuardi Sukisman"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 2", PICName: "Agus Sugiana"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 2", PICName: "Imam Taufik"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 2", PICName: "Gumilang Rizky Tresnapratama ST"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 2", PICName: "Arif Setiaji"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 2", PICName: "Irwan Suwandi"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 2", PICName: "Abdullah Farauk"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 2", PICName: "Nadya Audyra Rahardy"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 3", PICName: "Ferdian Ade Jiwandono"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 3", PICName: "Desrizal Azhar Arief"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 3", PICName: "Adam Firdaus"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 3", PICName: "Abipraya Wibawa Jati"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 3", PICName: "M. Anwar Hidayat"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 3", PICName: "Nadya Audyra Rahardy"},
		{Area: "Area Produksi", Kawasan: "Produksi CMD 3", PICName: "Abdullah Farauk"},
		{Area: "Area Lab QC & Inline", Kawasan: "Quality Control", PICName: "Nur Rahmawati R"},
		{Area: "Area Lab QC & Inline", Kawasan: "Quality Control", PICName: "Akbar Wiguna"},
		{Area: "Area Lab QC & Inline", Kawasan: "Quality Control", PICName: "Rina Anggraeni"},
		{Area: "Area Lab QC & Inline", Kawasan: "Quality Control", PICName: "Dena Syah Johan"},
		{Area: "Area Lab QC & Inline", Kawasan: "Quality Control", PICName: "Faza Nur Faqih"},
		{Area: "Area Lab QC & Inline", Kawasan: "Quality Control", PICName: "Rahmat Derajat"},
		{Area: "Area Lab QC & Inline", Kawasan: "Quality Control", PICName: "Abdullah Farauk"},
		{Area: "Area Lab QC & Inline", Kawasan: "Quality Control", PICName: "Krisna Arya Wibowo"},
		{Area: "Area Lab QC & Inline", Kawasan: "Quality Control", PICName: "Nadya Audyra Rahardy"},
		{Area: "Area Lab QC & Inline", Kawasan: "Quality Control", PICName: "Cacu Mulyani Nurdani"},
		{Area: "Area Lab QC & Inline", Kawasan: "Quality Control", PICName: "Yerris Sukma Preemasgar"},
		{Area: "Area Lab QC & Inline", Kawasan: "Quality Control", PICName: "Novia Eka Pratiwi"},
		{Area: "Area Lab QC & Inline", Kawasan: "Quality Control", PICName: "Rachmawati Dewi Lestari"},
		{Area: "Area Gudang", Kawasan: "Warehouse RMPM", PICName: "Sumaidi"},
		{Area: "Area Gudang", Kawasan: "Warehouse RMPM", PICName: "Sumadi Firmansyah"},
		{Area: "Area Gudang", Kawasan: "Warehouse RMPM", PICName: "Abdullah Farauk"},
		{Area: "Area Gudang", Kawasan: "Warehouse RMPM", PICName: "Nadya Audyra Rahardy"},
		{Area: "Area Gudang", Kawasan: "Warehouse MS", PICName: "Bambang Eko Sutrisno"},
		{Area: "Area Gudang", Kawasan: "Warehouse MS", PICName: "Syah Reza"},
		{Area: "Area Gudang", Kawasan: "Warehouse MS", PICName: "Zemi Faisal"},
		{Area: "Area Gudang", Kawasan: "Warehouse MS", PICName: "Nadya Audyra Rahardy"},
		{Area: "Area Gudang", Kawasan: "Warehouse MS", PICName: "Abdullah Farauk"},
		{Area: "Area Luar", Kawasan: "Area Luar CMD", PICName: "Ir. Hesti Indriyati"},
		{Area: "Area Luar", Kawasan: "Area Luar CMD", PICName: "Aldora Reginald Natha"},
		{Area: "Area Luar", Kawasan: "Area Luar CMD", PICName: "Yogi Indra Pratama"},
		{Area: "Area Luar", Kawasan: "Area Luar CMD", PICName: "Abdullah Farauk"},
		{Area: "Area Luar", Kawasan: "Area Luar CMD", PICName: "Nadya Audyra Rahardy"},
		{Area: "Area Lab QC & Inline", Kawasan: "Lab R&I", PICName: "Egi Fadlika"},
		{Area: "Area Lab QC & Inline", Kawasan: "Lab R&I", PICName: "Ike Mustikawati"},
		{Area: "Area Lab QC & Inline", Kawasan: "Lab R&I", PICName: "Duma Damaisanti Audri A"},
		{Area: "Area Lab QC & Inline", Kawasan: "Lab R&I", PICName: "Daniel Andrew Rahardjo"},
		{Area: "Area Lab QC & Inline", Kawasan: "Lab R&I", PICName: "Muhammad Hanif Fadhilah"},
		{Area: "Area Lab QC & Inline", Kawasan: "Lab R&I", PICName: "Alif Wahyu Faturahman"},
		{Area: "Area Lab QC & Inline", Kawasan: "Lab R&I", PICName: "Abdullah Farauk"},
		{Area: "Area Lab QC & Inline", Kawasan: "Lab R&I", PICName: "Nadya Audyra Rahardy"},
	}

	userMap := make(map[string]string) // PICName -> UserID
	for i, picDef := range picDefs {
		areaID := areaMap[picDef.Area]
		kawID := kawasanMap[picDef.Kawasan]
		if areaID == "" || kawID == "" {
			continue
		}

		// Check if user exists or create user record
		userID, exists := userMap[picDef.PICName]
		if !exists {
			var u authdomain.User
			err := db.Where("LOWER(\"FullName\") = LOWER(?)", picDef.PICName).First(&u).Error
			if err == nil {
				userID = u.UserID
			} else {
				// Generate clean username from full name
				cleanUsername := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(picDef.PICName, " ", "_"), ".", ""))
				cleanUsername = strings.ReplaceAll(cleanUsername, "__", "_")
				userID = fmt.Sprintf("USR-PIC-%03d", i+1)
				hashedPassword, _ := password.Hash("pic123")

				u = authdomain.User{
					UserID:       userID,
					DepartmentID: "DEPT-002",
					RoleID:       "ROLE-003", // Auditee / PIC
					PlantID:      &defaultPlantID,
					Username:     cleanUsername,
					FullName:     picDef.PICName,
					Email:        fmt.Sprintf("%s@cimory.com", cleanUsername),
					PasswordHash: hashedPassword,
					UserStatus:   authdomain.UserStatusActive,
				}
				db.Where("\"Username\" = ?", u.Username).FirstOrCreate(&u)
				userID = u.UserID
			}
			userMap[picDef.PICName] = userID
		}

		// Create PIC Mapping entry
		pmID := fmt.Sprintf("PICMAP-%03d", i+1)
		picMap := picdomain.PICMapping{
			PICMapID:    pmID,
			AreaID:      &areaID,
			KawasanID:   kawID,
			UserID:      userID,
			KategoriPIC: "Primary PIC",
		}
		db.Where("\"KawasanID\" = ? AND \"UserID\" = ?", kawID, userID).FirstOrCreate(&picMap)
	}

	// ── 7. Uraian Master ───────────────────────────────────────────────────────
	uraianTexts := []string{
		"Karyawan dalam kondisi sehat saat bekerja",
		"Karyawan tidak mengenakan perhiasan seperti kalung, cincin, gelang dan aksesoris lainnya",
		"Karyawan mengenakan pakaian kerja dan APD sesuai dengan yang telah ditentukan",
		"Semua karyawan mengikuti aturan standar yang telah ditentukan",
		"Dilarang membawa makanan dan minuman selain air mineral ke area laboratorium",
		"Terdapat sign petunjuk standar cuci tangan",
		"Saluran pembuangan lancar, tidak tersumbat dan bocor",
		"Keran air berfungsi dengan baik",
		"Kondisi wastafel terawat dan tidak menjadi sumber hama",
		"Fasilitas cuci dan sanitasi tersedia (sabun, tissue / dryer)",
		"Kontrol akses berfungsi dengan baik",
		"Penyimpanan barang sesuai dengan requirement",
		"Struktur bangunan kokoh dan dijaga dalam kondisi baik",
		"Tidak ada cat dinding yang mengelupas",
		"Pertemuan antara dinding dengan lantai membentuk sudut melengkung (curving)",
		"Permukaan bagian atap halus, tidak berlubang dan tidak ada sarang laba-laba",
		"Jendela / pintu dari material kaca dilapisi plastik film",
		"Pintu membuka ke arah luar",
		"Lampu dalam kondisi berfungsi, bersih dan tercover",
		"Bebas dari banjir, bau tidak sedap, debu, dan hama",
		"Jalan di dalam sarana area pabrik telah dikeraskan",
		"Mempunyai saluran pembuangan air yang baik",
		"Tidak ada genangan air",
		"Penyediaan tempat sampah yang memadai",
		"Cat warna pipa tidak terkelupas",
		"Terdapat akses masuk berbeda antara rute personil dan material",
		"Dilarang tidur di area pengolahan",
		"Palet plastik dalam kondisi baik dan tidak gompal",
		"Alat kebersihan diberikan identitas sesuai zona peruntukan",
		"Exaust secara visual bersih dan berfungsi",
	}

	// Link uraian items evenly across existing Details
	var allDetails []masterdomain.Detail
	db.Find(&allDetails)

	if len(allDetails) > 0 {
		for i, text := range uraianTexts {
			det := allDetails[i%len(allDetails)]
			uID := fmt.Sprintf("URN-REAL-%03d", i+1)
			uraian := masterdomain.Uraian{
				UraianID:      uID,
				DetailID:      det.DetailID,
				UraianText:    text,
				StandardScore: 100,
			}
			db.Where("\"UraianID\" = ? OR (LOWER(\"UraianText\") = LOWER(?) AND \"DetailID\" = ?)", uID, text, det.DetailID).
				Assign(masterdomain.Uraian{UraianID: uID, DetailID: det.DetailID, UraianText: text, StandardScore: 100}).
				FirstOrCreate(&uraian)
		}
	}

	// ── 8. Sample Inspection & Issue Seeding for Development ────────────────
	now := time.Now()
	dueDate := now.AddDate(0, 0, 7)

	sampleInsps := []struct {
		InspID   string
		AreaID   string
		KawID    string
		DKawID   string
		ResID    string
		UraianID string
		IssueID  string
		PICID    string
		Status   string
		Ket      string
	}{
		{
			InspID:   "INSP-REAL-001",
			AreaID:   "AREA-PRODUKSI",
			KawID:    "KWS-001",
			DKawID:   "DKWS-001",
			ResID:    "RES-REAL-001",
			UraianID: "URN-REAL-015",
			IssueID:  "ISSUE-REAL-001",
			PICID:    "USR-ADMIN-SENTUL",
			Status:   "Open",
			Ket:      "Pertemuan dinding dan lantai di area Proses CMD 1 belum melengkung (curving)",
		},
		{
			InspID:   "INSP-REAL-002",
			AreaID:   "AREA-PRODUKSI",
			KawID:    "KWS-001",
			DKawID:   "DKWS-002",
			ResID:    "RES-REAL-002",
			UraianID: "URN-REAL-020",
			IssueID:  "ISSUE-REAL-002",
			PICID:    "USR-PIC-001",
			Status:   "PendingValidation",
			Ket:      "Terdapat debu dan sisa material pada area filling",
		},
		{
			InspID:   "INSP-REAL-003",
			AreaID:   "AREA-GUDANG",
			KawID:    "KWS-005",
			DKawID:   "DKWS-016",
			ResID:    "RES-REAL-003",
			UraianID: "URN-REAL-029",
			IssueID:  "ISSUE-REAL-003",
			PICID:    "USR-PIC-002",
			Status:   "Open",
			Ket:      "Alat kebersihan di Gudang RMPM belum diberi label zona peruntukan",
		},
	}

	for _, s := range sampleInsps {
		inspHeader := inspectiondomain.InspectionHeader{
			InspectionID:              s.InspID,
			AreaID:                    s.AreaID,
			KawasanID:                 s.KawID,
			DetailKawasanID:           s.DKawID,
			InspectorID:               "USR-ADMIN-SENTUL",
			InspectionHeaderStatus:    inspectiondomain.InspectionStatusCompleted,
			InspectionHeaderCreatedAt: now,
		}
		db.Where("\"InspectionID\" = ?", s.InspID).FirstOrCreate(&inspHeader)

		inspResult := inspectiondomain.InspectionResult{
			ResultID:     s.ResID,
			InspectionID: s.InspID,
			UraianID:     s.UraianID,
			Checking:     "NG",
			Nilai:        0,
			Keterangan:   s.Ket,
		}
		db.Where("\"ResultID\" = ?", s.ResID).FirstOrCreate(&inspResult)

		iss := issuedomain.Issue{
			IssueID:        s.IssueID,
			ResultID:       s.ResID,
			IssuePICUserID: s.PICID,
			DueDate:        &dueDate,
			IssueStatus:    issuedomain.IssueStatus(s.Status),
			Keterangan:     s.Ket,
		}
		db.Where("\"IssueID\" = ?", s.IssueID).FirstOrCreate(&iss)
	}

	log.Println("✅ Real Master Data & Sample Issues seeded successfully.")
}
