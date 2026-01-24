const fontFamily = "JetBrains Mono";
const fontUrl = "https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600;700&display=swap";

let loaded = false;

export const loadFonts = async () => {
  if (loaded) return;

  // Load Google Font
  const link = document.createElement("link");
  link.href = fontUrl;
  link.rel = "stylesheet";
  document.head.appendChild(link);

  // Wait for font to load
  await document.fonts.ready;
  loaded = true;
};

export const FONT_FAMILY = `"JetBrains Mono", "SF Mono", "Monaco", "Consolas", monospace`;
