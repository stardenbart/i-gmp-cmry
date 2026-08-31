import { Pill } from "./Pill";
import { ROLE_STYLE, type PanduanRole } from "./theme";

export function RoleBadge({ role, label }: { role: PanduanRole; label?: string }) {
  const style = ROLE_STYLE[role];
  return <Pill style={label ? { ...style, label } : style} />;
}
