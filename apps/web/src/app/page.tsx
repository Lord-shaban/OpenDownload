import { Workspace } from "@/components/workspace";
import { cookies } from "next/headers";
export default async function Page() {
  return (
    <Workspace
      initialDark={(await cookies()).get("od_theme")?.value === "dark"}
    />
  );
}
