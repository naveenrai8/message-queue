package com.nr.messagequeuesql;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.transaction.annotation.EnableTransactionManagement;

@SpringBootApplication
@EnableTransactionManagement
public class MessageQueueSqlApplication {

    public static void main(String[] args) {
        SpringApplication.run(MessageQueueSqlApplication.class, args);
    }

}
