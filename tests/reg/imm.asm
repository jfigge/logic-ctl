  .org $0000

  .org $0200
  ldx #$2
  ldy #$fe
  dex
  dex
  dex
  inx
  inx
  inx
  iny
  iny
  iny
  dey
  dey
  dey

  ldx #$80
  lda #$00
  txa

  lda #$00
  ldx #$80
  tax

  ldy #$80
  lda #$00
  tya

  lda #$00
  ldy #$80
  tay
