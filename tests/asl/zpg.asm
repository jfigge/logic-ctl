  .org $0000
  .org $0200
;  lda #$24
;  sta $00
;  lda #$a2
;  sta $01
  asl $00
  asl $00
  asl $00
  asl $01
  asl $01
  asl $01

  .org $0000
  .byte $24
  .byte $a2
