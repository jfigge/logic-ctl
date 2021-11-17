  .org $0000
  .org $0200
  sec
  lda #$ff
  sbc #$03
  sbc #$fc
  sbc #$00
  sbc #$1
  sbc #$fe

  ; overflow test
  lda #$02
  sbc #$80
