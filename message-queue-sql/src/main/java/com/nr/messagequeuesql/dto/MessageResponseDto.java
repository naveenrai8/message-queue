package com.nr.messagequeuesql.dto;

import lombok.Builder;

import java.time.LocalDateTime;
import java.util.UUID;

@Builder
public record MessageResponseDto(
        String messages,
        UUID messageId,
        LocalDateTime leaseExpiredAt
) {
}
