import { LucideIcon } from "lucide-react";
import {
  Building2,
  MapPin,
  Layers,
  ListChecks,
  ClipboardList,
  FileText,
  CheckSquare,
} from "lucide-react";

export interface MasterTabConfig {
  id: string;
  label: string;
  icon: LucideIcon;
  endpoint: string;
}

export const MASTER_TABS: MasterTabConfig[] = [
  { id: "departments", label: "Department",   icon: Building2,    endpoint: "/master/departments" },
  { id: "areas",       label: "Area",         icon: MapPin,       endpoint: "/master/area" },
  { id: "kawasans",    label: "Kawasan",      icon: Layers,       endpoint: "/master/kawasan" },
  { id: "detail-kawasans", label: "Detail Kawasan", icon: ListChecks, endpoint: "/master/detail-kawasan" },
  { id: "aspeks",      label: "Aspek Audit",    icon: ClipboardList,endpoint: "/master/aspek" },
  { id: "details",     label: "Detail Audit",   icon: FileText,     endpoint: "/master/details" },
  { id: "urains",      label: "Uraian",         icon: CheckSquare,  endpoint: "/master/urain" },
];
