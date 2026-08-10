import { LucideIcon, Tag } from "lucide-react";
import {
  Factory,
  Building2,
  MapPin,
  Layers,
  ListChecks,
  ClipboardList,
  FileText,
  CheckSquare,
  Activity,
  Wrench,
  Warehouse,
} from "lucide-react";

export interface MasterTabConfig {
  id: string;
  label: string;
  icon: LucideIcon;
  endpoint: string;
}

export const MASTER_TABS: MasterTabConfig[] = [
  { id: "plants",          label: "Plant (Pabrik)", icon: Factory,      endpoint: "/master/plants" },
  { id: "departments",     label: "Department",     icon: Building2,    endpoint: "/master/departments" },
  { id: "areas",           label: "Area",           icon: MapPin,       endpoint: "/master/area" },
  { id: "kawasans",        label: "Kawasan",        icon: Layers,       endpoint: "/master/kawasan" },
  { id: "detail-kawasans", label: "Detail Kawasan", icon: ListChecks,   endpoint: "/master/detail-kawasan" },
  { id: "aspeks",          label: "Aspek Audit",    icon: ClipboardList,endpoint: "/master/aspek" },
  { id: "details",         label: "Detail Audit",   icon: FileText,     endpoint: "/master/details" },
  { id: "urains",          label: "Uraian",         icon: CheckSquare,  endpoint: "/master/urain" },
  { id: "hei",            label: "Klasifikasi HEI",icon: Tag,          endpoint: "/master/hei" },
  { id: "habits",          label: "Habit",          icon: Activity,     endpoint: "/master/habits" },
  { id: "equipments",      label: "Equipment",      icon: Wrench,       endpoint: "/master/equipments" },
  { id: "infrastructures",  label: "Infrastructure", icon: Warehouse,    endpoint: "/master/infrastructures" },
];
