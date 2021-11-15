  .org $0000
  .org $0200
  lda #$40
  ldy #$02
  sta ($30),y
  lda #$80
  ldy #$22
  sta ($32),y

  .org $0030
  .word $021e
  .word $01fe
