  .org $0000
  .org $0200
  ldx #$10
  lda #$24
  sta $10
  lda #$a2
  sta $11
  asl $00,x
  asl $00,x
  asl $00,x
  inx
  asl $00,x
  asl $00,x
  asl $00,x
