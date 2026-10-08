export const DEFAULT_THEME = "light";

/**
 * Runs inline in <head> before the first paint so a saved dark (or system
 * dark) preference does not flash the light theme while React hydrates.
 * Lives outside the "use client" provider so the server layout gets the
 * string itself, not a client reference.
 */
export const THEME_INIT_SCRIPT = `try{var t=localStorage.getItem("theme")||"${DEFAULT_THEME}";if(t==="system")t=window.matchMedia("(prefers-color-scheme: dark)").matches?"dark":"light";document.documentElement.classList.toggle("dark",t==="dark")}catch(e){}`;
