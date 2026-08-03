import os
import sys
from reportlab.lib.pagesizes import letter
from reportlab.platypus import (
    SimpleDocTemplate, Paragraph, Spacer, Table, TableStyle, PageBreak, KeepTogether, HRFlowable
)
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib import colors
from reportlab.pdfgen import canvas

class NumberedCanvas(canvas.Canvas):
    def __init__(self, *args, **kwargs):
        super(NumberedCanvas, self).__init__(*args, **kwargs)
        self._saved_page_states = []

    def showPage(self):
        self._saved_page_states.append(dict(self.__dict__))
        self._startPage()

    def save(self):
        num_pages = len(self._saved_page_states)
        for state in self._saved_page_states:
            self.__dict__.update(state)
            self.draw_page_number(num_pages)
            super(NumberedCanvas, self).showPage()
        super(NumberedCanvas, self).save()

    def draw_page_number(self, page_count):
        self.saveState()
        self.setFont("Helvetica-Bold", 8)
        self.setFillColor(colors.HexColor("#64748B"))
        
        # Header (Top of Page)
        self.drawString(54, 750, "PRANOR PLATFORM ENTERPRISE TECHNICAL ARCHITECTURE MANUAL")
        self.drawRightString(612 - 54, 750, "PRANOR-SPEC-2026-V10")
        self.setStrokeColor(colors.HexColor("#CBD5E1"))
        self.setLineWidth(0.5)
        self.line(54, 742, 612 - 54, 742)

        # Footer (Bottom of Page)
        self.line(54, 50, 612 - 54, 50)
        self.setFont("Helvetica", 8)
        self.drawString(54, 38, "Confidential — For Enterprise Architect & Technical Review Only")
        page_text = f"Page {self._pageNumber} of {page_count}"
        self.drawRightString(612 - 54, 38, page_text)
        self.restoreState()

def build_pdf(md_file_path, pdf_file_path):
    doc = SimpleDocTemplate(
        pdf_file_path,
        pagesize=letter,
        leftMargin=54,
        rightMargin=54,
        topMargin=72,
        bottomMargin=72
    )

    styles = getSampleStyleSheet()

    # Custom Color Palette
    PRIMARY = colors.HexColor("#0F172A")    # Dark Navy
    SECONDARY = colors.HexColor("#0284C7")  # Sky Blue
    ACCENT = colors.HexColor("#0D9488")     # Teal
    TEXT_DARK = colors.HexColor("#1E293B")  # Slate 800
    BG_LIGHT = colors.HexColor("#F8FAFC")   # Light Slate

    # Custom Typography Styles
    style_title = ParagraphStyle(
        'DocTitle',
        parent=styles['Normal'],
        fontName='Helvetica-Bold',
        fontSize=18,
        leading=22,
        textColor=colors.white,
        spaceAfter=6
    )
    style_subtitle = ParagraphStyle(
        'DocSubtitle',
        parent=styles['Normal'],
        fontName='Helvetica-Bold',
        fontSize=9,
        leading=12,
        textColor=colors.HexColor("#38BDF8")
    )
    style_h1 = ParagraphStyle(
        'Heading1_Custom',
        parent=styles['Normal'],
        fontName='Helvetica-Bold',
        fontSize=13,
        leading=17,
        textColor=PRIMARY,
        spaceBefore=14,
        spaceAfter=6,
        keepWithNext=True
    )
    style_h2 = ParagraphStyle(
        'Heading2_Custom',
        parent=styles['Normal'],
        fontName='Helvetica-Bold',
        fontSize=10,
        leading=14,
        textColor=SECONDARY,
        spaceBefore=10,
        spaceAfter=4,
        keepWithNext=True
    )
    style_body = ParagraphStyle(
        'Body_Custom',
        parent=styles['Normal'],
        fontName='Helvetica',
        fontSize=8.5,
        leading=12,
        textColor=TEXT_DARK,
        spaceAfter=5
    )
    style_bullet = ParagraphStyle(
        'Bullet_Custom',
        parent=styles['Normal'],
        fontName='Helvetica',
        fontSize=8.5,
        leading=12,
        textColor=TEXT_DARK,
        leftIndent=12,
        spaceAfter=3
    )
    style_code = ParagraphStyle(
        'Code_Custom',
        parent=styles['Normal'],
        fontName='Courier',
        fontSize=7.5,
        leading=9.5,
        textColor=colors.HexColor("#38BDF8"),
        spaceAfter=1
    )

    story = []

    # Document Header Title Block Box
    header_content = [
        [Paragraph("Pranor Vault — Technical Architecture Reference Manual", style_title)],
        [Paragraph("Reference: PRANOR-SPEC-VAULT-2026-V10  |  Classification: Technical Manual  |  Phases 1-87", style_subtitle)]
    ]
    header_table = Table(header_content, colWidths=[504])
    header_table.setStyle(TableStyle([
        ('BACKGROUND', (0,0), (-1,-1), PRIMARY),
        ('PADDING', (0,0), (-1,-1), 12),
        ('VALIGN', (0,0), (-1,-1), 'MIDDLE'),
        ('BOTTOMPADDING', (0,0), (-1,0), 2),
    ]))
    story.append(header_table)
    story.append(Spacer(1, 12))

    # Read Markdown content
    with open(md_file_path, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    in_code_block = False
    code_lines = []

    for line in lines:
        raw_line = line.rstrip('\n')

        # Code block handler
        if raw_line.startswith('```'):
            if in_code_block:
                # Close code block
                in_code_block = False
                code_table_data = [[Paragraph(cline, style_code)] for cline in code_lines]
                if code_table_data:
                    c_table = Table(code_table_data, colWidths=[504])
                    c_table.setStyle(TableStyle([
                        ('BACKGROUND', (0,0), (-1,-1), colors.HexColor("#0F172A")),
                        ('PADDING', (0,0), (-1,-1), 3),
                        ('LEFTPADDING', (0,0), (-1,-1), 8),
                    ]))
                    story.append(c_table)
                    story.append(Spacer(1, 6))
                code_lines = []
            else:
                in_code_block = True
                code_lines = []
            continue

        if in_code_block:
            code_lines.append(raw_line.replace('<', '&lt;').replace('>', '&gt;'))
            continue

        # Headers
        if raw_line.startswith('## '):
            story.append(Paragraph(raw_line.replace('## ', ''), style_h1))
            story.append(HRFlowable(width="100%", thickness=1, color=SECONDARY, spaceBefore=2, spaceAfter=4))
        elif raw_line.startswith('### '):
            story.append(Paragraph(raw_line.replace('### ', ''), style_h2))
        elif raw_line.startswith('- ') or raw_line.startswith('* '):
            item_text = raw_line[2:].replace('**', '<b>', 1).replace('**', '</b>', 1)
            story.append(Paragraph(f"• {item_text}", style_bullet))
        elif raw_line.startswith('1. ') or raw_line.startswith('2. ') or raw_line.startswith('3. ') or raw_line.startswith('4. '):
            item_text = raw_line[3:].replace('**', '<b>', 1).replace('**', '</b>', 1)
            story.append(Paragraph(f"{raw_line[:3]} {item_text}", style_bullet))
        elif len(raw_line.strip()) > 0:
            formatted_text = raw_line.replace('**', '<b>', 1).replace('**', '</b>', 1)
            story.append(Paragraph(formatted_text, style_body))

    doc.build(story, canvasmaker=NumberedCanvas)
    print(f"Technical Architecture PDF rendered successfully: {pdf_file_path}")

if __name__ == '__main__':
    md_path = sys.argv[1] if len(sys.argv) > 1 else "/home/developer/workspace/pranor/docs/modules/vault.md"
    pdf_path = sys.argv[2] if len(sys.argv) > 2 else "/home/developer/workspace/pranor/docs/modules/vault.pdf"
    build_pdf(md_path, pdf_path)
