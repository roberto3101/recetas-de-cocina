const fs = require("fs");
const { Document, Packer, Paragraph, TextRun, HeadingLevel } = require("docx");

const doc = new Document({
  creator: "test",
  styles: {
    default: { document: { run: { font: "Calibri", size: 22 } } },
    paragraphStyles: [
      { id: "Heading1", name: "Heading 1", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 36, bold: true, font: "Calibri", color: "8B3A2A" },
        paragraph: { spacing: { before: 300, after: 200 }, outlineLevel: 0 } },
    ],
  },
  sections: [{
    properties: { page: { size: { width: 12240, height: 15840 }, margin: { top: 1440, right: 1440, bottom: 1440, left: 1440 } } },
    children: [
      new Paragraph({ heading: HeadingLevel.HEADING_1, children: [new TextRun({ text: "Hola Word" })] }),
      new Paragraph({ children: [new TextRun({ text: "Esto es un parrafo simple." })] }),
    ],
  }],
});

Packer.toBuffer(doc).then((buf) => {
  fs.writeFileSync(__dirname + "/test-minimal.docx", buf);
  console.log("OK", buf.length, "bytes");
});
