const sharp = require("sharp");
const fs = require("fs");
const path = require("path");

const dirPublic = path.join(__dirname, "..", "public");
const svg = fs.readFileSync(path.join(dirPublic, "icon.svg"));

async function generar() {
  await sharp(svg).resize(192, 192).png().toFile(path.join(dirPublic, "icon-192.png"));
  await sharp(svg).resize(512, 512).png().toFile(path.join(dirPublic, "icon-512.png"));
  // Maskable: padding interno (safe area) para que el icon no se recorte en superficies redondas
  await sharp({
    create: { width: 512, height: 512, channels: 4, background: "#8b3a2a" },
  })
    .composite([{ input: await sharp(svg).resize(360, 360).png().toBuffer(), gravity: "center" }])
    .png()
    .toFile(path.join(dirPublic, "icon-maskable-512.png"));
  await sharp(svg).resize(180, 180).png().toFile(path.join(dirPublic, "apple-touch-icon.png"));
  console.log("OK iconos generados");
}

generar().catch((e) => { console.error(e); process.exit(1); });
