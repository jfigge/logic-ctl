  .org $0000
  .org $0200
;  ldx $00
;  lda #$24
;  sta $00
;  lda #$a2
;  sta $01
  asl $00,x
  asl $00,x
  asl $00,x
  inx
  asl $01,x
  asl $01,x
  asl $01,x

  .org $0000
  .byte $24
  .byte $a2
